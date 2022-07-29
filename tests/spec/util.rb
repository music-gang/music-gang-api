# frozen_string_literal: true

require 'faker'

def username
  Faker::Internet.username
end

def email
  Faker::Internet.email name: username, separators: '.'
end

def password
  'Password123!'
end

def new_user
  User.new username, email, password
end
