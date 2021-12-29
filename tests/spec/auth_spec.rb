# frozen_string_literal: true

require 'faker'

require_relative '../services/lib'

container = service_container

name = Faker::Name.name
email = Faker::Internet.email name: name, separators: '.'
password = 'Password123!'

user = User.new name, email, password

describe 'Flow Auth: ' do
  describe 'register a new user' do
    context 'given correct data' do
      it 'returns token pairs' do
        container.auth_service.register user
      end
    end

    context 'given incorrect data' do
      context 'for example an empty name' do
        it 'returns an error' do
          name = ''

          begin
            container.auth_service.register User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          end
        end
      end

      context 'for example an incorrect password' do
        it 'returns an error' do
          name = Faker::Name.name
          email = Faker::Internet.email name: name, separators: '.'
          password = 'non-valid-password'

          begin
            container.auth_service.register User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          end
        end
      end

      context 'for example an incorrect email' do
        it 'returns an error' do
          name = Faker::Name.name
          email = 'incorrect-email'

          begin
            container.auth_service.register User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          end
        end
      end
    end
  end
end
