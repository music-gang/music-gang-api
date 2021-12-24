# frozen_string_literal: true

require 'uri'
require 'net/http'

# ServiceHTTP
class ServiceHTTP
  attr_accessor :url, :schema

  def initialize(schema, url)
    @schema = schema
    @url = url
  end

  def base_url
    "#{@schema}://#{@url}"
  end

  def endpoint(url)
    "#{base_url}/#{url}"
  end
end
