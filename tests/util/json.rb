# frozen_string_literal: true

require 'json'

# Jsonizable module is an util to implements json serialization for a concrete class
module Jsonizable
  def to_hash(*_args)
    raise NotImplementedError
  end

  def to_json(*_args)
    to_hash.to_json
  end
end
