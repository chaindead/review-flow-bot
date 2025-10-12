# GitLab Rails Console script to create test users with access tokens
# Run this inside the GitLab container: docker exec -i gitlab-local gitlab-rails console < create-test-users.rb

users_data = [
  {
    username: 'developer1',
    name: 'Developer One',
    email: 'dev1@example.com',
    password: 'O8WSQYPOpl3KObrzn78e',
    token: 'dev1_token_12345678'
  },
  {
    username: 'developer2',
    name: 'Developer Two',
    email: 'dev2@example.com',
    password: 'O8WSQYPOpl3KObrzn78e',
    token: 'dev2_token_12345678'
  },
  {
    username: 'reviewer1',
    name: 'Reviewer One',
    email: 'rev1@example.com',
    password: 'O8WSQYPOpl3KObrzn78e',
    token: 'rev1_token_12345678'
  },
  {
    username: 'reviewer2',
    name: 'Reviewer Two',
    email: 'rev2@example.com',
    password: 'O8WSQYPOpl3KObrzn78e',
    token: 'rev2_token_12345678'
  }
]

created_tokens = []

users_data.each do |user_data|
  puts "Creating user: #{user_data[:username]}..."
  
  user = User.find_by_username(user_data[:username])
  
  if user
    puts "  User #{user_data[:username]} already exists"
  else
    # Create user with proper namespace setup
    user = User.new(
      username: user_data[:username],
      name: user_data[:name],
      email: user_data[:email],
      password: user_data[:password],
      password_confirmation: user_data[:password]
    )
    user.skip_confirmation!
    
    # Build namespace manually
    user.build_namespace(
      name: user_data[:username],
      path: user_data[:username],
      organization: Organizations::Organization.default_organization
    )
    
    if user.save
      puts "  Successfully created user: #{user_data[:username]}"
    else
      puts "  Failed to create user: #{user_data[:username]}"
      puts "  Errors: #{user.errors.full_messages.join(', ')}"
      next
    end
  end
  
  # Create personal access token
  puts "  Creating access token for #{user_data[:username]}..."
  
  existing_token = user.personal_access_tokens.find_by(name: 'API Token')
  
  if existing_token
    puts "  Token already exists for #{user_data[:username]}"
    created_tokens << {
      username: user_data[:username],
      token: existing_token.token || user_data[:token],
      email: user_data[:email]
    }
  else
    begin
      token = user.personal_access_tokens.create!(
        scopes: ['read_user', 'read_api', 'api'],
        name: 'API Token',
        expires_at: 365.days.from_now
      )
      token.set_token(user_data[:token])
      token.save!
      
      puts "  Successfully created token for #{user_data[:username]}"
      created_tokens << {
        username: user_data[:username],
        token: user_data[:token],
        email: user_data[:email]
      }
    rescue => e
      puts "  Failed to create token: #{e.message}"
    end
  end
end

puts "\n" + "="*70
puts "User Creation Complete"
puts "="*70
puts "\nCreated users with credentials:"
users_data.each do |user_data|
  puts "  - #{user_data[:username]} / #{user_data[:password]} (#{user_data[:email]})"
end

puts "\n" + "="*70
puts "Access Tokens (use for GitLab API):"
puts "="*70
created_tokens.each do |token_data|
  puts "\n#{token_data[:username]}:"
  puts "  Token: #{token_data[:token]}"
  puts "  Email: #{token_data[:email]}"
end

puts "\n" + "="*70
puts "Test token with curl:"
puts "  curl -H 'PRIVATE-TOKEN: dev1_token_12345678' http://localhost:8929/api/v4/user"
puts "="*70

