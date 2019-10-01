require_relative '../utils'

RSpec.describe Isutrain::Utils do
  class App
    include Isutrain::Utils
  end

  describe '.get_usable_train_class_list' do
    let(:from_station) { nil }

    subject { App.new.get_usable_train_class_list(from_station, from_station).length }

    context '全部止まる' do
      let(:from_station) do
        {
          'id' => 1,
          'name' => '全部止まる',
          'distance' => 10.0,
          'is_stop_express' => true,
          'is_stop_semi_express' => true,
          'is_stop_local' => true,
        }
      end

      it do
        is_expected.to eq(3)
      end
    end

    context 'ちょっと止まる' do
      let(:from_station) do
        {
          'id' => 1,
          'name' => 'ちょっと止まる',
          'distance' => 10.0,
          'is_stop_express' => false,
          'is_stop_semi_express' => true,
          'is_stop_local' => true,
        }
      end

      it do
        is_expected.to eq(2)
      end
    end

    context '各駅' do
      let(:from_station) do
        {
          'id' => 1,
          'name' => '各駅',
          'distance' => 10.0,
          'is_stop_express' => false,
          'is_stop_semi_express' => false,
          'is_stop_local' => true,
        }
      end

      it do
        is_expected.to eq(1)
      end
    end
  end
end
