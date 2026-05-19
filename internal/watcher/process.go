package watcher

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	client "gitlab.com/gitlab-org/api/client-go"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/db/models"
	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/gitlab"
	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/lang"
)

func (w *Watcher) Watch() {
	ctx := contextFromChan(w.stopChan)

	for {
		select {
		case <-w.stopChan:
			return
		case <-time.After(w.cfg.PollInterval):
			var mrs []models.MergeRequest
			if err := w.db.DB().NewSelect().
				Model(&mrs).
				Scan(ctx); err != nil {
				log.Err(err).Msg("get mr from db")
				continue
			}

			for _, mr := range mrs {
				select {
				case <-w.stopChan:
					return
				default:
				}

				l := log.Logger.With().Str("url", mr.Link).Logger()
				if err := w.processOne(ctx, l, mr); err != nil {
					log.Err(err).Msg("process mr")
					continue
				}
			}
		}
	}
}

func (w *Watcher) processOne(ctx context.Context, l zerolog.Logger, mr models.MergeRequest) error {
	mr = models.MergeRequest{ProjectID: mr.ProjectID, MRID: mr.MRID}
	if err := w.db.DB().NewSelect().
		Model(&mr).
		WherePK().
		Relation("Author").
		Relation("Reviewers.Reviewer").
		Scan(ctx); err != nil {
		return errors.Wrap(err, "scan mr")
	}

	// NEED ACTION? SECTION
	ds, descUpdatedAt, err := w.git.GetDiscussion(ctx, mr.ProjectID, mr.MRID)
	if err != nil {
		return errors.Wrap(err, "list discussions")
	}

	// PROCESS SECTION: MERGE REQ STATUS
	glMr, err := w.git.GetMergeRequest(ctx, mr.ProjectID, mr.MRID)
	if err != nil {
		return errors.Wrap(err, "get MR")
	}

	if glMr.State == gitlab.StatusLocked {
		l.Info().Msg("skip locked MR")
		return nil
	}

	if glMr.State == gitlab.StatusClosed || glMr.State == gitlab.StatusMerged {
		_, err = w.db.DB().NewDelete().Model(&mr).WherePK().Exec(ctx)
		if err != nil {
			return errors.Wrap(err, "delete mr")
		}

		// send notification: merged/closed
		l.Info().Str("status", glMr.State).Msg("notify delete mr")
		return w.tg.Send(mr.AuthorID, "watch.notify."+glMr.State, lang.Args{"MR": mr})
	}

	if glMr.State != gitlab.StatusOpened {
		l.Warn().Str("status", glMr.State).Msg("skip due unknown status")
		return nil
	}

	if glMr.UpdatedAt != nil && glMr.UpdatedAt.After(descUpdatedAt) {
		descUpdatedAt = *glMr.UpdatedAt
	}

	if descUpdatedAt.IsZero() || descUpdatedAt.Equal(mr.GitlabUpdatedAt) {
		l.Debug().
			Time("gl-updated", descUpdatedAt).
			Msg("skip due no changes")
		return nil
	}

	l.Debug().
		Times("cur/new", []time.Time{mr.GitlabUpdatedAt, descUpdatedAt}).
		Msg("new update: stating process")

	// PROCESS SECTION: COMMENTS
	notifyNoteAuthor := []string{}
	notifyNoteReviewer := make(map[int][]string)
	authorID := int(mr.Author.GitlabID)

	resolveMap := make(map[int]bool)
	for _, r := range mr.Reviewers {
		resolveMap[int(r.Reviewer.GitlabID)] = true
	}

	for _, d := range ds {
		ps := discussionParticipants(d.Notes, authorID)
		for _, n := range d.Notes {
			if n.System {
				continue
			}

			if n.Resolvable {
				for _, p := range ps {
					resolveMap[p] = resolveMap[p] && n.Resolved
				}
			}

			if n.CreatedAt == nil {
				l.Warn().Interface("note", n).Msg("skip due no updated note")
				continue
			}

			if !mr.GitlabUpdatedAt.Before(*n.CreatedAt) {
				l.Trace().Interface("comment", n.Body).Msg("skip not due already seen")
				continue
			}

			log.Debug().Str("comment", n.Body).
				Time("updated", *n.CreatedAt).
				Time("compare-with-after", mr.GitlabUpdatedAt).
				Time("updated-to-apply", descUpdatedAt).
				Msg("new comment")

			if n.Author.ID != authorID {
				notifyNoteAuthor = append(notifyNoteAuthor, n.Body)
				continue
			}

			for _, p := range ps {
				notifyNoteReviewer[p] = append(notifyNoteReviewer[p], n.Body)
			}
		}
	}

	// PROCESS SECTION: APPROVALS
	approvals, _, err := w.git.C().MergeRequests.GetMergeRequestApprovals(mr.ProjectID, mr.MRID)
	if err != nil {
		return errors.Wrap(err, "get approvals")
	}

	approveMap := approvalsToMap(approvals)

	// UPDATE SECTION: MERGE REQ
	mr.GitlabUpdatedAt = descUpdatedAt
	mr.Title = glMr.Title
	mr.Link = glMr.WebURL
	if _, err = w.db.DB().NewUpdate().Model(&mr).WherePK().Exec(ctx); err != nil {
		return errors.Wrap(err, "update mr")
	}

	// UPDATE SECTION: REVIEWER
	statusChanges := make(map[int64]*models.MRReviewer)
	for _, r := range mr.Reviewers {
		var status string
		switch {
		case approveMap[r.Reviewer.GitlabID]:
			status = models.MRStatusApproved
		case !resolveMap[int(r.Reviewer.GitlabID)]:
			status = models.MRStatusRejected
		default:
			status = models.MRStatusPending
		}

		if status == r.Status {
			l.Debug().Interface("reviewer", r.Reviewer).Msg("skip reviewer update due same status")
			continue
		}

		r.LastReminderAt = time.Now()
		r.Status = status
		if _, err = w.db.DB().NewUpdate().Model(r).WherePK().Exec(ctx); err != nil {
			return errors.Wrap(err, "update mr")
		}

		statusChanges[r.Reviewer.GitlabID] = r
	}

	l.Debug().
		Interface("StatusChange", statusChanges).
		Interface("notifyNoteAuthor", notifyNoteAuthor).
		Interface("notifyNoteReviewer", notifyNoteReviewer).
		Interface("approveMap", approveMap).
		Interface("resolveMap", resolveMap).
		//Interface("ds", ds).
		//Interface("approvals", approvals).
		//Interface("glMr", glMr).
		//Interface("mr", mr).
		Send()

	// NOTIFICATION SECTION
	for _, r := range mr.Reviewers {
		comments := notifyNoteReviewer[int(r.Reviewer.GitlabID)]
		statusChange := statusChanges[r.Reviewer.GitlabID]
		if len(comments) == 0 && !changesIn(statusChanges, models.MRStatusPending) {
			l.Debug().Str("reviewer", r.Reviewer.TelegramUsername).Msg("skip reviewer notification due empty")
			continue
		}

		// NOTIFY REVIEWER: status -> pending
		args := lang.Args{
			"MR":           mr,
			"StatusChange": statusChange,
			"Comments":     comments,
		}
		l.Info().Interface("args", args).Str("uname", r.Reviewer.TelegramUsername).Msg("notify reviewer")

		if err = w.tg.Send(r.Reviewer.ID, "watch.notify.reviewer", args); err != nil {
			l.Err(err).Str("reviewer", r.Reviewer.TelegramUsername).Msg("send reviewer notification")
		}
	}

	if !changesIn(statusChanges, models.MRStatusApproved, models.MRStatusRejected) && len(notifyNoteAuthor) == 0 {
		l.Debug().Msg("skip author notification due empty")
		return nil
	}

	// NOTIFY AUTHOR: status -> rejected, approved
	args := lang.Args{
		"MR":            mr,
		"StatusChanges": statusChanges,
		"Comments":      notifyNoteAuthor,
	}
	l.Info().Interface("args", args).Msg("notify author")

	if err = w.tg.Send(mr.AuthorID, "watch.notify.author", args); err != nil {
		return errors.Wrap(err, "send author notification")
	}

	return nil
}

func discussionParticipants(ns []*client.Note, authorID int) []int {
	m := make(map[int]bool)
	for _, n := range ns {
		if !n.Resolvable {
			continue
		}

		if n.Author.ID == authorID {
			continue
		}

		m[n.Author.ID] = true
	}

	ps := make([]int, 0, len(m))
	for p := range m {
		ps = append(ps, p)
	}

	return ps
}

func approvalsToMap(a *client.MergeRequestApprovals) map[int64]bool {
	m := make(map[int64]bool, len(a.Approvers))

	for _, u := range a.ApprovedBy {
		m[int64(u.User.ID)] = true
	}

	return m
}

func changesIn(changes map[int64]*models.MRReviewer, statuses ...string) bool {
	ss := make(map[string]bool, len(statuses))
	for _, s := range statuses {
		ss[s] = true
	}

	for _, change := range changes {
		if change != nil && ss[change.Status] {
			return true
		}
	}

	return false
}
