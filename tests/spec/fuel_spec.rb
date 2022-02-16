# frozen_string_literal: true

require_relative '../services/lib'

container = service_container

describe 'Fuel flow:' do
  describe 'stats' do
    context 'calling the stats service' do
      it 'returns a FuelStat' do
        fuel_stat = container.fuel_service.stats
        expect(fuel_stat).to be_a FuelStat

        expect(fuel_stat.fuel_capacity).to be_a_kind_of Integer
        expect(fuel_stat.fuel_used).to be_a_kind_of Integer
        expect(fuel_stat.last_refuel_amount).to be_a_kind_of Integer

        expect(fuel_stat.last_refuel_at).to be_a Time
      end
    end
  end
end
