# frozen_string_literal: true

# User entity class
class User
  include Jsonizable

  attr_accessor :name, :email
  attr_reader :created_at, :updated_at, :password

  def initialize(name, email, password)
    @name = name
    @email = email
    @password = password
    @created_at = Time.now.utc
    @updated_at = Time.now.utc
  end

  def to_hash
    {
      name: @name,
      email: @email,
      password: @password,
      created_at: @created_at,
      updated_at: @updated_at
    }
  end
end
