# frozen_string_literal: true

# Revision entity.
class Revision
  attr_accessor :id, :created_at, :rev, :version, :contract_id, :notes, :max_fuel

  def initialize(id: nil,
                 created_at: nil,
                 rev: nil,
                 version: nil,
                 contract_id: nil,
                 notes: nil,
                 max_fuel: nil)
    @id = id
    @created_at = created_at
    @rev = rev
    @version = version
    @contract_id = contract_id
    @notes = notes
    @max_fuel = max_fuel
  end

  class << self
    # from_hash
    # @param hash [Hash]
    def from_hash(hash)
      validate_hash hash
      Revision.new id: hash[:id], created_at: Time.parse(hash[:created_at]), rev: hash[:rev], version: hash[:version], contract_id: hash[:contract_id], notes: hash[:notes], max_fuel: hash[:max_fuel]
    end

    # Validate hash.
    # @param hash [Hash]
    def validate_hash(hash)
      raise 'missing id' unless hash.key? :id
      raise 'missing created at' unless hash.key? :created_at
      raise 'missing rev' unless hash.key? :rev
      raise 'missing version' unless hash.key? :version
      raise 'missing contract id' unless hash.key? :contract_id
      raise 'missing notes' unless hash.key? :notes
      raise 'missing max fuel' unless hash.key? :max_fuel
    end
  end
end
