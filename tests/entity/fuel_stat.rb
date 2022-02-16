# frozen_string_literal: true

# FuelStat is a class that represents a fuel stat.
class FuelStat
  include Jsonizable

  attr_accessor :fuel_capacity, :fuel_used, :last_refuel_amount, :last_refuel_at

  def initialize(fuel_capacity, fuel_used, last_refuel_amount, last_refuel_at)
    @fuel_capacity = fuel_capacity
    @fuel_used = fuel_used
    @last_refuel_amount = last_refuel_amount
    @last_refuel_at = last_refuel_at
  end

  def to_hash
    {
      fuel_capacity: @fuel_capacity,
      fuel_used: @fuel_used,
      last_refuel_amount: @last_refuel_amount,
      last_refuel_at: @last_refuel_at
    }
  end

  class << self
    def from_hash(hash)
      validate_hash hash

      FuelStat.new hash[:fuel_capacity], hash[:fuel_used], hash[:last_refuel_amount], Time.parse(hash[:last_refuel_at])
    end

    def validate_hash(hash)
      raise 'missing fuel capacity' unless hash.key? :fuel_capacity
      raise 'missing fuel used' unless hash.key? :fuel_used
      raise 'missing last refuel amount' unless hash.key? :last_refuel_amount

      raise 'missing last refuel at' unless hash.key? :last_refuel_at

      raise 'invalid last refuel at' unless hash[:last_refuel_at].is_a? String
    end
  end
end
