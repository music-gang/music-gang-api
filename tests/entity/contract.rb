# frozen_string_literal: true

# Contract entity.
class Contract
  include Jsonizable

  attr_accessor :id, :name, :description, :user_id, :visibility, :max_fuel, :stateful, :created_at, :updated_at

  def initialize(id: nil,
                 name: nil,
                 description: nil,
                 user_id: nil,
                 visibility: nil,
                 max_fuel: nil,
                 stateful: false,
                 created_at: nil,
                 updated_at: nil)
    @id = id
    @name = name
    @description = description
    @user_id = user_id
    @visibility = visibility
    @max_fuel = max_fuel
    @stateful = stateful
    @created_at = created_at
    @updated_at = updated_at
  end

  def to_hash
    {
      id: @id,
      name: @name,
      description: @description,
      user_id: @user_id,
      visibility: @visibility,
      max_fuel: @max_fuel,
      stateful: @stateful,
      created_at: @created_at,
      updated_at: @updated_at
    }
  end

  class << self
    def from_hash(hash)
      validate_hash hash

      Contract.new id: hash[:id], name: hash[:name], description: hash[:description], user_id: hash[:user_id], visibility: hash[:visibility], max_fuel: hash[:max_fuel], stateful: hash[:stateful], created_at: Time.parse(hash[:created_at]), updated_at: Time.parse(hash[:updated_at])
    end

    def validate_hash(hash)
      raise 'missing id' unless hash.key? :id
      raise 'missing name' unless hash.key? :name
      raise 'missing description' unless hash.key? :description
      raise 'missing user id' unless hash.key? :user_id
      raise 'missing visibility' unless hash.key? :visibility
      raise 'missing max fuel' unless hash.key? :max_fuel
      raise 'missing stateful' unless hash.key? :stateful
      raise 'missing created at' unless hash.key? :created_at
      raise 'missing updated at' unless hash.key? :updated_at
    end
  end
end
