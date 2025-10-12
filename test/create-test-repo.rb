# GitLab Rails Console script to create a test repository with merge request
# Run this inside the GitLab container: docker exec -i gitlab-local gitlab-rails console < create-test-repo.rb

require 'securerandom'

# Configuration
GITLAB_URL = ENV['GITLAB_URL'] || 'http://localhost:8929'
timestamp = Time.now.strftime('%Y%m%d_%H%M%S')
PROJECT_NAME = "test-project-#{timestamp}"
PROJECT_PATH = "test-project-#{timestamp}"

puts "="*70
puts "Creating Test Repository and Merge Request"
puts "="*70
puts ""

# Find root user (project owner)
root_user = User.find_by_username('root')
if root_user.nil?
  puts "ERROR: Root user not found!"
  exit 1
end

# Find all test users
test_users = []
['developer1', 'developer2', 'reviewer1', 'reviewer2'].each do |username|
  user = User.find_by_username(username)
  if user
    test_users << user
    puts "Found user: #{username}"
  else
    puts "WARNING: User #{username} not found"
  end
end

if test_users.empty?
  puts "ERROR: No test users found. Run create-test-users.rb first!"
  exit 1
end

puts ""
puts "Creating project: #{PROJECT_NAME}..."

# Create project
project = Projects::CreateService.new(
  root_user,
  name: PROJECT_NAME,
  path: PROJECT_PATH,
  namespace_id: root_user.namespace.id,
  visibility_level: Gitlab::VisibilityLevel::PRIVATE,
  initialize_with_readme: true
).execute

unless project.persisted?
  puts "ERROR: Failed to create project"
  puts "Errors: #{project.errors.full_messages.join(', ')}"
  exit 1
end

puts "✓ Project created successfully"
puts "  Project ID: #{project.id}"
puts ""

# Add test users as project members
puts "Adding users to project..."
test_users.each do |user|
  # Access level: 40 = Maintainer, 30 = Developer, 20 = Reporter
  member = project.add_developer(user)
  if member.persisted?
    puts "  ✓ Added #{user.username} as Developer"
  else
    puts "  ✗ Failed to add #{user.username}"
  end
end

puts ""
puts "Creating branches and commits for merge request..."

begin
  # Get repository
  repository = project.repository
  
  # Create a feature branch
  feature_branch_name = 'feature/test-feature'
  
  # First, ensure we have the default branch
  default_branch = project.default_branch || 'main'
  
  # Create feature branch from default branch
  result = Branches::CreateService.new(project, root_user).execute(
    feature_branch_name,
    default_branch
  )
  
  if result[:status] == :success
    puts "  ✓ Created branch: #{feature_branch_name}"
  else
    puts "  ✗ Failed to create branch: #{result[:message]}"
    # Try to continue anyway in case branch exists
  end
  
  # Create a commit on the feature branch
  file_path = 'test-file.txt'
  file_content = "This is a test file\nCreated at: #{Time.now}\nRandom: #{SecureRandom.hex(8)}"
  
  commit_result = Files::CreateService.new(
    project,
    root_user,
    commit_message: 'Add test file for merge request',
    start_branch: feature_branch_name,
    branch_name: feature_branch_name,
    file_path: file_path,
    file_content: file_content
  ).execute
  
  if commit_result[:status] == :success
    puts "  ✓ Created commit on #{feature_branch_name}"
  else
    puts "  ✗ Failed to create commit: #{commit_result[:message]}"
  end
  
  # Create another commit
  file_path2 = 'README.md'
  file_content2 = "# #{PROJECT_NAME}\n\nTest project for review bot testing.\n\nCreated: #{Time.now}"
  
  # Check if README exists, if so update it, otherwise create it
  existing_file = repository.blob_at(feature_branch_name, file_path2)
  
  if existing_file
    commit_result2 = Files::UpdateService.new(
      project,
      root_user,
      commit_message: 'Update README',
      start_branch: feature_branch_name,
      branch_name: feature_branch_name,
      file_path: file_path2,
      file_content: file_content2,
      last_commit_id: repository.commit(feature_branch_name).id
    ).execute
  else
    commit_result2 = Files::CreateService.new(
      project,
      root_user,
      commit_message: 'Update README',
      start_branch: feature_branch_name,
      branch_name: feature_branch_name,
      file_path: file_path2,
      file_content: file_content2
    ).execute
  end
  
  if commit_result2[:status] == :success
    puts "  ✓ Created second commit"
  end
  
  puts ""
  puts "Creating merge request..."
  
  # Create merge request
  mr_params = {
    title: "Test MR: Add test feature",
    description: "This is a test merge request for bot testing.\n\nCreated at: #{Time.now}",
    source_branch: feature_branch_name,
    target_branch: default_branch,
    author: root_user
  }
  
  merge_request = MergeRequests::CreateService.new(
    project: project,
    current_user: root_user,
    params: mr_params
  ).execute
  
  if merge_request.persisted?
    puts "  ✓ Merge request created successfully"
    puts "  MR IID: #{merge_request.iid}"
  else
    puts "  ✗ Failed to create merge request"
    puts "  Errors: #{merge_request.errors.full_messages.join(', ')}"
  end
  
rescue => e
  puts "ERROR during branch/commit creation: #{e.message}"
  puts e.backtrace.first(5)
end

puts ""
puts "="*70
puts "Test Environment Created Successfully!"
puts "="*70
puts ""
puts "Project Information:"
puts "  Name: #{PROJECT_NAME}"
puts "  Path: #{project.full_path}"
puts "  ID: #{project.id}"
puts ""
puts "Links:"
puts "  Project URL: #{GITLAB_URL}/#{project.full_path}"
puts "  MR URL: #{GITLAB_URL}/#{project.full_path}/-/merge_requests/#{merge_request&.iid || '1'}"
puts ""
puts "Project Members:"
test_users.each do |user|
  puts "  - #{user.username} (#{user.email})"
end
puts ""
puts "Commands to test:"
puts "  # Get project info"
puts "  curl -H 'PRIVATE-TOKEN: dev1_token_12345678' #{GITLAB_URL}/api/v4/projects/#{project.id}"
puts ""
puts "  # Get merge request"
puts "  curl -H 'PRIVATE-TOKEN: dev1_token_12345678' #{GITLAB_URL}/api/v4/projects/#{project.id}/merge_requests/#{merge_request&.iid || '1'}"
puts ""
puts "="*70

