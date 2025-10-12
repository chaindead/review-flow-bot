# Review Flow Bot

Review Flow Bot streamlines the code review process by automatically managing reviewer assignments and notifications. 
It integrates with GitLab to track merge requests and uses Telegram for communication.

## Installation

### From Releases

1. Download the latest release for your platform
2. Extract the binary to your desired location
3. Make the binary executable:
   ```bash
   chmod +x review-flow-bot
   ```

## Getting Started
> Use `--help` flag to see all availible flags

1. Create a Telegram bot via [@BotFather](https://t.me/botfather)
2. Obtain a GitLab personal access token with `read_api` scope
3. Configure environment variables with help of `review-flow-bot --envs`
4. Start the bot with `review-flow-bot`

## Usage
> Use `/help` command in bot to see all available commands

* Users can authenticate using `/login <gitlab_token>` with `read_user` scope.
* Admins can assign users to teams using `/assign @username team_name`
* Team members can request reviews using `/review <mr_url>`
* Bot monitors active merge requests and
    * will notify the interested user if there are changes in the review status
    * allows user to view the status of your reviews via /my
