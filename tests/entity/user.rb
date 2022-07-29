# frozen_string_literal: true

# User entity class
class User
  include Jsonizable

  attr_accessor :name, :email, :token_pairs, :id
  attr_reader :created_at, :updated_at, :password

  def initialize(name, email, password)
    @id = 0
    @name = name
    @email = email
    @password = password
    @created_at = Time.now.utc
    @updated_at = Time.now.utc

    # @type [TokenPair]
    @token_pairs = nil
  end

  def to_hash
    {
      id: @id,
      name: @name,
      email: @email,
      password: @password,
      created_at: @created_at,
      updated_at: @updated_at
    }
  end

  class << self
    def from_hash(hash)
      user = User.new hash[:name], hash[:email], hash[:password]
      user.id = hash[:id]
      user
    end
  end
end
