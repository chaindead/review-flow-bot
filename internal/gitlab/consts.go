package gitlab

const (
	StatusOpened = "opened"
	StatusClosed = "closed"
	StatusLocked = "locked"
	StatusMerged = "merged"

	NoteSimple     = ""               // in main, simple write, not resolvable
	NoteDiff       = "DiffNote"       // in diff resolvable
	NoteDiscussion = "DiscussionNote" // // in main, resolvable on review req change
)
