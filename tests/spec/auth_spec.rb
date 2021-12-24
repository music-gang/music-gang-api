# frozen_string_literal: true

require 'faker'

require_relative '../services/lib'

container = service_container

name = Faker::Name.name
email = Faker::Internet.email name, '.'
password = 'Password123!'

user = User.new name, email, password

RSpec.describe 'Flow Auth: ' do
  describe 'register a new user' do
    context 'given correct data' do
      it 'returns token pairs' do
        container.auth_service.register user
      end
    end
  end
end
