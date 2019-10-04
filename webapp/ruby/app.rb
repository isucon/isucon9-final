require 'json'
require 'openssl'
require 'uri'
require 'net/http'
require 'securerandom'
require 'sinatra/base'
require 'mysql2'
require 'mysql2-cs-bind'

require './utils'

module Isutrain
  class App < Sinatra::Base
    include Utils

    class Error < StandardError; end
    class ErrorNoRows < StandardError; end

    configure :development do
      require 'sinatra/reloader'

      register Sinatra::Reloader
      also_reload './utils.rb'
    end

    set :protection, false
    set :show_exceptions, false
    set :session_secret, 'tagomoris'
    set :sessions, key: 'session_isutrain', expire_after: 3600

    helpers do
      def db
        Thread.current[:db] ||= Mysql2::Client.new(
          host: ENV['MYSQL_HOSTNAME'] || '127.0.0.1',
          port: ENV['MYSQL_PORT'] || '3306',
          database: ENV['MYSQL_USER'] || 'isutrain',
          username: ENV['MYSQL_DATABASE'] || 'isutrain',
          password: ENV['MYSQL_PASSWORD'] || 'isutrain',
          charset: 'utf8mb4',
          database_timezone: :local,
          cast_booleans: true,
          symbolize_keys: true,
          reconnect: true,
        )
      end

      def get_user
        user_id = session[:user_id]

        return nil, 401, 'no session' if user_id.nil?

        user = db.xquery(
          'SELECT * FROM `users` WHERE `id` = ?',
          user_id,
        ).first

        return nil, 401, "user not found #{user_id}" if user.nil?

        [user, 200, '']
      end

      def get_distance_fare(orig_to_dest_distance)
        distance_fare_list = db.query(
          'SELECT `distance`, `fare` FROM `distance_fare_master` ORDER BY `distance`',
        )

        last_distance = 0.0
        last_fare = 0

        distance_fare_list.each do |distance_fare|
          puts "#{orig_to_dest_distance} #{distance_fare[:distance]} #{distance_fare[:fare]}"

          break if last_distance < orig_to_dest_distance && orig_to_dest_distance < distance_fare[:distance]

          last_distance = distance_fare[:distance]
          last_fare = distance_fare[:fare]
        end

        last_fare
      end

      def fare_calc(date, dep_station, dest_station, train_class, seat_class)
        # 料金計算メモ
        # 距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)

        from_station = db.xquery(
          'SELECT * FROM `station_master` WHERE `id` = ?',
          dep_station,
        ).first

        raise ErrorNoRows if from_station.nil?

        to_station = db.xquery(
          'SELECT * FROM `station_master` WHERE `id` = ?',
          dest_station,
        ).first

        raise ErrorNoRows if to_station.nil?

        puts "distance #{(to_station[:distance] - from_station[:distance]).abs}"

        dist_fare = get_distance_fare((to_station[:distance] - from_station[:distance]).abs)
        puts "distFare #{dist_fare}"

        # 期間・車両・座席クラス倍率
        fare_list = db.xquery(
          'SELECT * FROM `fare_master` WHERE `train_class` = ? AND `seat_class` = ? ORDER BY `start_date`',
          train_class,
          seat_class,
        )

        raise Error, 'fare_master does not exists' if fare_list.to_a.length.zero?

        selected_fare = fare_list.first

        date = Date.new(date.year, date.month, date.day)
        fare_list.each do |fare|
          start_date = Date.new(fare[:start_date].year, fare[:start_date].month, fare[:start_date].day)

          if start_date <= date
            puts "#{fare[:start_date]} #{fare[:fare_multiplier]}"
            selected_fare = fare
          end
        end

        puts '%%%%%%%%%%%%%%%%%%%'

        (dist_fare * selected_fare[:fare_multiplier]).floor
      end

      def make_reservation_response(reservation)
        departure = db.xquery(
          'SELECT `departure` FROM `train_timetable_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ? AND `station` = ?',
          reservation[:date].strftime('%Y/%m/%d'),
          reservation[:train_class],
          reservation[:train_name],
          reservation[:departure],
          cast: false,
        ).first

        raise ErrorNoRows, 'departure is not found' if departure.nil?

        arrival = db.xquery(
          'SELECT `arrival` FROM `train_timetable_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ? AND `station` = ?',
          reservation[:date].strftime('%Y/%m/%d'),
          reservation[:train_class],
          reservation[:train_name],
          reservation[:arrival],
          cast: false,
        ).first

        raise ErrorNoRows, 'arrival is not found' if arrival.nil?

        reservation_response = {
          reservation_id: reservation[:reservation_id],
          date: reservation[:date].strftime('%Y/%m/%d'),
          amount: reservation[:amount],
          adult: reservation[:adult],
          child: reservation[:child],
          departure: reservation[:departure],
          arrival: reservation[:arrival],
          train_class: reservation[:train_class],
          train_name: reservation[:train_name],
          departure_time: departure[:departure],
          arrival_time: arrival[:arrival],
        }

        reservation_response[:seats] = db.xquery(
          'SELECT * FROM `seat_reservations` WHERE `reservation_id` = ?',
          reservation[:reservation_id],
        ).to_a

        # 1つの予約内で車両番号は全席同じ
        reservation_response[:car_number] = reservation_response[:seats].first[:car_number]

        if reservation_response[:seats].first[:car_number] == 0
          reservation_response[:seat_class] = 'non-reserved'
        else
          seat = db.xquery(
            'SELECT * FROM `seat_master` WHERE `train_class` = ? AND `car_number` = ? AND `seat_column` = ? AND `seat_row` = ?',
            reservation[:train_class],
            reservation_response[:car_number],
            reservation_response[:seats].first[:seat_column],
            reservation_response[:seats].first[:seat_row],
          ).first

          raise ErrorNoRows, 'seat is not found' if seat.nil?

          reservation_response[:seat_class] = seat[:seat_class]
        end

        reservation_response[:seats].each do |v|
          # omit
          v[:reservation_id] = 0
          v[:car_number] = 0
        end

        reservation_response
      end

      def body_params
        @body_params ||= JSON.parse(request.body.tap(&:rewind).read, symbolize_names: true)
      end

      def message_response(message)
        content_type :json

        {
          is_error: false,
          message: message,
        }.to_json
      end

      def halt_with_error(status = 500, message = 'unknown')
        headers = {
          'Content-Type' => 'application/json',
        }
        response = {
          is_error: true,
          message: message,
        }

        halt status, headers, response.to_json
      end
    end

    post '/initialize' do
      db.query('TRUNCATE seat_reservations')
      db.query('TRUNCATE reservations')
      db.query('TRUNCATE users')

      content_type :json
      {
        available_days: AVAILABLE_DAYS,
        language: 'ruby',
      }.to_json
    end

    get '/api/settings' do
      payment_api = ENV['PAYMENT_API'] || 'http://127.0.0.1:5000'

      content_type :json
      { payment_api: payment_api }.to_json
    end

    get '/api/stations' do
      stations = db.query('SELECT * FROM `station_master` ORDER BY `id`').map do |station|
        station.slice(:id, :name, :is_stop_express, :is_stop_semi_express, :is_stop_local)
      end

      content_type :json
      stations.to_json
    end

    get '/api/train/search' do
      date = Time.iso8601(params[:use_at]).getlocal

      halt_with_error 404, '予約可能期間外です' unless check_available_date(date)

      from_station = db.xquery(
        'SELECT * FROM station_master WHERE name = ?',
        params[:from],
      ).first

      if from_station.nil?
        puts 'fromStation: no rows'
        halt_with_error 400, 'fromStation: no rows'
      end

      to_station = db.xquery(
        'SELECT * FROM station_master WHERE name = ?',
        params[:to],
      ).first

      if to_station.nil?
        puts 'toStation: no rows'
        halt_with_error 400, 'toStation: no rows'
      end

      is_nobori = from_station[:distance] > to_station[:distance]

      usable_train_class_list = get_usable_train_class_list(from_station, to_station)

      train_list = if params[:train_class].nil? || params[:train_class].empty?
        db.xquery(
          'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` IN (?) AND `is_nobori` = ?',
          date.strftime('%Y/%m/%d'),
          usable_train_class_list,
          is_nobori,
        )
      else
        db.xquery(
          'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` IN (?) AND `is_nobori` = ? AND `train_class` = ?',
          date.strftime('%Y/%m/%d'),
          usable_train_class_list,
          is_nobori,
          params[:train_class],
        )
      end

      stations = db.xquery(
        "SELECT * FROM `station_master` ORDER BY `distance` #{is_nobori ? 'DESC' : 'ASC'}",
      )

      puts "From #{from_station}"
      puts "To #{to_station}"

      train_search_response_list = []

      train_list.each do |train|
        is_seeked_to_first_station = false
        is_contains_origin_station = false
        is_contains_dest_station = false
        i = 0

        stations.each do |station|
          unless is_seeked_to_first_station
            # 駅リストを列車の発駅まで読み飛ばして頭出しをする
            # 列車の発駅以前は止まらないので無視して良い
            if station[:name] == train[:start_station]
              is_seeked_to_first_station = true
            else
              next
            end
          end

          if station[:id] == from_station[:id]
            # 発駅を経路中に持つ編成の場合フラグを立てる
            is_contains_origin_station = true
          end

          if station[:id] == to_station[:id]
            if is_contains_origin_station
              # 発駅と着駅を経路中に持つ編成の場合
              is_contains_dest_station = true
            else
              # 出発駅より先に終点が見つかったとき
              puts 'なんかおかしい'
            end

            break
          end

          if station[:name] == train[:last_station]
            # 駅が見つからないまま当該編成の終点に着いてしまったとき
            break
          end

          i += 1
        end

        if is_contains_origin_station && is_contains_dest_station
          # 列車情報

          departure = db.xquery(
            'SELECT `departure` FROM `train_timetable_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ? AND `station` = ?',
            date.strftime('%Y/%m/%d'),
            train[:train_class],
            train[:train_name],
            from_station[:name],
            cast: false,
          ).first

          departure_date = Time.parse("#{date.strftime('%Y/%m/%d')} #{departure[:departure]} +09:00 JST")

          next unless date < departure_date

          arrival = db.xquery(
            'SELECT `arrival` FROM `train_timetable_master` WHERE date = ? AND `train_class` = ? AND `train_name` = ? AND `station` = ?',
            date.strftime('%Y/%m/%d'),
            train[:train_class],
            train[:train_name],
            to_station[:name],
            cast: false,
          ).first

          premium_avail_seats = get_available_seats(train, from_station, to_station, 'premium', false)
          premium_smoke_avail_seats = get_available_seats(train, from_station, to_station, 'premium', true)
          reserved_avail_seats = get_available_seats(train, from_station, to_station, 'reserved', false)
          reserved_smoke_avail_seats = get_available_seats(train, from_station, to_station, 'reserved', true)

          premium_avail = '○'
          if premium_avail_seats.length.zero?
            premium_avail = '×'
          elsif premium_avail_seats.length < 10
            premium_avail = '△'
          end

          premium_smoke_avail = '○'
          if premium_smoke_avail_seats.length.zero?
            premium_smoke_avail = '×'
          elsif premium_smoke_avail_seats.length < 10
            premium_smoke_avail = '△'
          end

          reserved_avail = '○'
          if reserved_avail_seats.length.zero?
            reserved_avail = '×'
          elsif reserved_avail_seats.length < 10
            reserved_avail = '△'
          end

          reserved_smoke_avail = '○'
          if reserved_smoke_avail_seats.length.zero?
            reserved_smoke_avail = '×'
          elsif reserved_smoke_avail_seats.length < 10
            reserved_smoke_avail = '△'
          end

          # 空席情報
          seat_availability = {
            premium: premium_avail,
            premium_smoke: premium_smoke_avail,
            reserved: reserved_avail,
            reserved_smoke: reserved_smoke_avail,
            non_reserved: '○',
          }

          # 料金計算
          premium_fare = fare_calc(date, from_station[:id], to_station[:id], train[:train_class], 'premium')
          premium_fare = premium_fare * params[:adult].to_i + premium_fare / 2 * params[:child].to_i

          reserved_fare = fare_calc(date, from_station[:id], to_station[:id], train[:train_class], 'reserved')
          reserved_fare = reserved_fare * params[:adult].to_i + reserved_fare / 2 * params[:child].to_i

          non_reserved_fare = fare_calc(date, from_station[:id], to_station[:id], train[:train_class], 'non-reserved')
          non_reserved_fare = non_reserved_fare * params[:adult].to_i + non_reserved_fare / 2 * params[:child].to_i

          fare_information = {
            premium: premium_fare,
            premium_smoke: premium_fare,
            reserved: reserved_fare,
            reserved_smoke: reserved_fare,
            non_reserved: non_reserved_fare,
          }

          train_search_response = {
            train_class: train[:train_class],
            train_name: train[:train_name],
            start: train[:start_station],
            last: train[:last_station],
            departure: from_station[:name],
            arrival: to_station[:name],
            departure_time: departure[:departure],
            arrival_time: arrival[:arrival],
            seat_availability: seat_availability,
            seat_fare: fare_information,
          }

          train_search_response_list << train_search_response

          break if train_search_response_list.length >= 10
        end
      end

      content_type :json
      train_search_response_list.to_json
    end

    get '/api/train/seats' do
      date = Time.iso8601(params[:date]).getlocal

      halt_with_error 404, '予約可能期間外です' unless check_available_date(date)

      train = db.xquery(
        'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ?',
        date.strftime('%Y/%m/%d'),
        params[:train_class],
        params[:train_name],
      ).first

      halt_with_error 404, '列車が存在しません' if train.nil?

      from_name = params[:from]
      from_station = db.xquery(
        'SELECT * FROM `station_master` WHERE `name` = ?',
        from_name,
      ).first

      if from_station.nil?
        puts 'fromStation: no rows'
        halt_with_error 400, 'fromStation: no rows'
      end

      to_name = params[:to]
      to_station = db.xquery(
        'SELECT * FROM `station_master` WHERE `name` = ?',
        to_name,
      ).first

      if to_station.nil?
        puts 'toStation: no rows'
        halt_with_error 400, 'toStation: no rows'
      end

      usable_train_class_list = get_usable_train_class_list(from_station, to_station)
      unless usable_train_class_list.include?(train[:train_class])
        puts 'invalid train_class'
        halt_with_error 400, 'invalid train_class'
      end

      seat_list = db.xquery(
        'SELECT * FROM `seat_master` WHERE `train_class` = ? AND `car_number` = ? ORDER BY `seat_row`, `seat_column`',
        params[:train_class],
        params[:car_number],
      )

      seat_information_list = []

      seat_list.each do |seat|
        s = {
          row: seat[:seat_row],
          column: seat[:seat_column],
          class: seat[:seat_class],
          is_smoking_seat: seat[:is_smoking_seat],
          is_occupied: false
        }

        query = <<__EOF
          SELECT
            `s`.*
          FROM
            `seat_reservations` `s`,
            `reservations` `r`
          WHERE
            `r`.`date` = ? AND
            `r`.`train_class` = ? AND
            `r`.`train_name` = ? AND
            `car_number` = ? AND
            `seat_row` = ? AND
            `seat_column` = ?
__EOF

        seat_reservation_list = db.xquery(
          query,
          date.strftime('%Y/%m/%d'),
          seat[:train_class],
          params[:train_name],
          seat[:car_number],
          seat[:seat_row],
          seat[:seat_column],
        )

        p seat_reservation_list

        seat_reservation_list.each do |seat_reservation|
          reservation = db.xquery(
            'SELECT * FROM `reservations` WHERE `reservation_id` = ?',
            seat_reservation[:reservation_id],
          ).first

          departure_station = db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            reservation[:departure],
          ).first

          arrival_station = db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            reservation[:arrival],
          ).first

          if train[:is_nobori]
            # 上り
            if to_station[:id] < arrival_station[:id] && from_station[:id] <= arrival_station[:id]
              # pass
            elsif to_station[:id] >= departure_station[:id] && from_station[:id] > departure_station[:id]
              # pass
            else
              s[:is_occupied] = true
            end
          else
            # 下り
            if from_station[:id] < departure_station[:id] && to_station[:id] <= departure_station[:id]
              # pass
            elsif from_station[:id] >= arrival_station[:id] && to_station[:id] > arrival_station[:id]
              # pass
            else
              s[:is_occupied] = true
            end
          end
        end

        puts s[:is_occupied] ? 'true' : 'false'

        seat_information_list << s
      end

      # 各号車の情報
      simple_car_information_list = []
      i = 1
      loop do
        seat = db.xquery(
          'SELECT * FROM `seat_master` WHERE `train_class` = ? AND `car_number` = ? ORDER BY `seat_row`, `seat_column` LIMIT 1',
          params[:train_class],
          i,
        ).first

        break if seat.nil?

        simple_car_information = {
          car_number: i,
          seat_class: seat[:seat_class],
        }

        simple_car_information_list << simple_car_information

        i += 1
      end

      c = {
        date: date.strftime('%Y/%m/%d'),
        train_class: params[:train_class],
        train_name: params[:train_name],
        car_number: params[:car_number].to_i,
        seats: seat_information_list,
        cars: simple_car_information_list,
      }

      content_type :json
      c.to_json
    end

    post '/api/train/reserve' do
      date = Time.iso8601(body_params[:date]).getlocal

      halt_with_error 404, '予約可能期間外です' unless check_available_date(date)

      db.query('BEGIN')

      begin
        tmas = begin
          db.xquery(
            'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ?',
            date.strftime('%Y/%m/%d'),
            body_params[:train_class],
            body_params[:train_name],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '列車データの取得に失敗しました'
        end

        if tmas.nil?
          db.query('ROLLBACK')
          halt_with_error 404, '列車データがみつかりません'
        end

        puts tmas

        # 列車自体の駅IDを求める
        departure_station = begin
          db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            tmas[:start_station],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, 'リクエストされた列車の始発駅データの取得に失敗しました'
        end

        if departure_station.nil?
          db.query('ROLLBACK')
          halt_with_error 404, 'リクエストされた列車の始発駅データがみつかりません'
        end

        # Arrive
        arrival_station = begin
          db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            tmas[:last_station],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, 'リクエストされた列車の終着駅データの取得に失敗しました'
        end

        if arrival_station.nil?
          db.query('ROLLBACK')
          halt_with_error 404, 'リクエストされた列車の終着駅データがみつかりません'
        end

        # From
        from_station = begin
          db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            body_params[:departure],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '乗車駅データの取得に失敗しました'
        end

        if from_station.nil?
          db.query('ROLLBACK')
          halt_with_error 404, "乗車駅データがみつかりません #{body_params[:departure]}"
        end

        # To
        to_station = begin
          db.xquery(
            'SELECT * FROM `station_master` WHERE `name` = ?',
            body_params[:arrival],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '降車駅駅データの取得に失敗しました'
        end

        if to_station.nil?
          db.query('ROLLBACK')
          halt_with_error 404, "降車駅駅データがみつかりません #{body_params[:arrival]}"
        end

        case body_params[:train_class]
        when '最速'
          if !from_station[:is_stop_express] || !to_station[:is_stop_express]
            db.query('ROLLBACK')
            halt_with_error 400, '最速の止まらない駅です'
          end
        when '中間'
          if !from_station[:is_stop_semi_express] || !to_station[:is_stop_semi_express]
            db.query('ROLLBACK')
            halt_with_error 400, '中間の止まらない駅です'
          end
        when '遅いやつ'
          if !from_station[:is_stop_local] || !to_station[:is_stop_local]
            db.query('ROLLBACK')
            halt_with_error 400, '遅いやつの止まらない駅です'
          end
        else
          db.query('ROLLBACK')
          halt_with_error 400, 'リクエストされた列車クラスが不明です'
        end

        # 運行していない区間を予約していないかチェックする
        if tmas[:is_nobori]
          if from_station[:id] > departure_station[:id] || to_station[:id] > departure_station[:id]
            db.query('ROLLBACK')
            halt_with_error 400, 'リクエストされた区間に列車が運行していない区間が含まれています'
          end

          if arrival_station[:id] >= from_station[:id] || arrival_station[:id] > to_station[:id]
            db.query('ROLLBACK')
            halt_with_error 400, 'リクエストされた区間に列車が運行していない区間が含まれています'
          end
        else
          if from_station[:id] < departure_station[:id] || to_station[:id] < departure_station[:id]
            db.query('ROLLBACK')
            halt_with_error 400, 'リクエストされた区間に列車が運行していない区間が含まれています'
          end

          if arrival_station[:id] <= from_station[:id] || arrival_station[:id] < to_station[:id]
            db.query('ROLLBACK')
            halt_with_error 400, 'リクエストされた区間に列車が運行していない区間が含まれています'
          end
        end

        # あいまい座席検索
        # seatsが空白の時に発動する
        if body_params[:seats].empty?
          if body_params[:seat_class] != 'non-reserved'
            train = begin
              db.xquery(
                'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ?',
                date.strftime('%Y/%m/%d'),
                body_params[:train_class],
                body_params[:train_name],
              ).first
            rescue Mysql2::Error => e
              db.query('ROLLBACK')
              puts e.message
              halt_with_error 400, e.message
            end

            if train.nil?
              db.query('ROLLBACK')
              halt_with_error 404, 'train is not found'
            end

            usable_train_class_list = get_usable_train_class_list(from_station, to_station)
            unless usable_train_class_list.include?(train[:train_class])
              err = 'invalid train_class'
              puts err
              db.query('ROLLBACK')
              halt_with_error 400, err
            end

            body_params[:seats] = [] # 座席リクエスト情報は空に
            (1..16).each do |carnum|
              seat_list = begin
                db.xquery(
                  'SELECT * FROM `seat_master` WHERE `train_class` = ? AND `car_number` = ? AND `seat_class` = ? AND `is_smoking_seat` = ? ORDER BY `seat_row`, `seat_column`',
                  body_params[:train_class],
                  carnum,
                  body_params[:seat_class],
                  !!body_params[:is_smoking_seat],
                )
              rescue Mysql2::Error => e
                db.query('ROLLBACK')
                puts e.message
                halt_with_error 400, e.message
              end

              seat_information_list = []
              seat_list.each do |seat|
                s = {
                  row: seat[:seat_row],
                  column: seat[:seat_column],
                  class: seat[:seat_class],
                  is_smoking_seat: seat[:is_smoking_seat],
                  is_occupied: false,
                }

                seat_reservation_list = begin
                  db.xquery(
                    'SELECT `s`.* FROM `seat_reservations` `s`, `reservations` `r` WHERE `r`.`date` = ? AND `r`.`train_class` = ? AND `r`.`train_name` = ? AND `car_number` = ? AND `seat_row` = ? AND `seat_column` = ? FOR UPDATE',
                    date.strftime('%Y/%m/%d'),
                    seat[:train_class],
                    body_params[:train_name],
                    seat[:car_number],
                    seat[:seat_row],
                    seat[:seat_column],
                  )
                rescue Mysql2::Error => e
                  db.query('ROLLBACK')
                  puts e.message
                  halt_with_error 400, e.message
                end

                seat_reservation_list.each do |seat_reservation|
                  reservation = begin
                    db.xquery(
                      'SELECT * FROM `reservations` WHERE `reservation_id` = ? FOR UPDATE',
                      seat_reservation[:reservation_id],
                    ).first
                  rescue Mysql2::Error => e
                    db.query('ROLLBACK')
                    puts e.message
                    halt_with_error 400, e.message
                  end

                  if reservation.nil?
                    db.query('ROLLBACK')
                    halt_with_error 404, 'reservation is not found'
                  end

                  departure_station = begin
                    db.xquery(
                      'SELECT * FROM `station_master` WHERE `name` = ?',
                      reservation[:departure],
                    ).first
                  rescue Mysql2::Error => e
                    db.query('ROLLBACK')
                    puts e.message
                    halt_with_error 400, e.message
                  end

                  if departure_station.nil?
                    db.query('ROLLBACK')
                    halt_with_error 404, 'departure_station is not found'
                  end

                  arrival_station = begin
                    db.xquery(
                      'SELECT * FROM `station_master` WHERE `name` = ?',
                      reservation[:arrival],
                    ).first
                  rescue Mysql2::Error => e
                    db.query('ROLLBACK')
                    puts e.message
                    halt_with_error 400, e.message
                  end

                  if arrival_station.nil?
                    db.query('ROLLBACK')
                    halt_with_error 404, 'arrival_station is not found'
                  end

                  if train[:is_nobori]
                    # 上り
                    if to_station[:id] < arrival_station[:id] && from_station[:id] <= arrival_station[:id]
                      # pass
                    elsif to_station[:id] >= departure_station[:id] && from_station[:id] > departure_station[:id]
                      # pass
                    else
                      s[:is_occupied] = true
                    end
                  else
                    # 下り
                    if from_station[:id] < departure_station[:id] && to_station[:id] <= departure_station[:id]
                      # pass
                    elsif from_station[:id] >= arrival_station[:id] && to_station[:id] > arrival_station[:id]
                      # pass
                    else
                      s[:is_occupied] = true
                    end
                  end
                end

                seat_information_list << s
              end

              # 曖昧予約席とその他の候補席を選出
              vague_seat = {}


              reserved = false
              vargue = true
              seatnum = body_params[:adult] + body_params[:child] - 1     # 全体の人数からあいまい指定席分を引いておく
              if body_params[:column].nil? || body_params[:column].empty? # A/B/C/D/Eを指定しなければ、空いている適当な指定席を取るあいまいモード
                seatnum = body_params[:adult] + body_params[:child]       # あいまい指定せず大人＋小人分の座席を取る
                reserved = true                                           # dummy
                vargue = false                                            # dummy
              end

              candidate_seats = []

              # シート分だけ回して予約できる席を検索
              i = 0
              seat_information_list.each do |seat|
                if seat[:column] == body_params[:column] && !seat[:is_occupied] && !reserved && vargue # あいまい席があいてる
                  vague_seat = seat
                  reserved = true
                elsif !seat[:is_occupied] && i < seatnum # 単に席があいてる
                  candidate_seats << {
                    row: seat[:row],
                    column: seat[:column],
                  }

                  i += 1
                end
              end

              if vargue && reserved
                body_params[:seats] << vague_seat
              end

              if i > 0
                body_params[:seats].concat(candidate_seats)
              end

              if body_params[:seats].length < body_params[:adult] + body_params[:child]
                # リクエストに対して席数が足りてない
                # 次の号車にうつしたい
                puts '-----------------'
                puts "現在検索中の車両: #{carnum}号車, リクエスト座席数: #{body_params[:adult] + body_params[:child]}, 予約できそうな座席数: #{body_params[:seats].length}, 不足数: #{body_params[:adult] + body_params[:child] - body_params[:seats].length}"
                puts 'リクエストに対して座席数が不足しているため、次の車両を検索します。'

                body_params[:seats] = []
                if carnum == 16
                  puts 'この新幹線にまとめて予約できる席数がなかったから検索をやめるよ'
                  break
                end
              end

              puts "空き実績: #{carnum}号車 シート: #{body_params[:seats]} 席数: #{body_params[:seats].length}"

              if body_params[:seats].length >= body_params[:adult] + body_params[:child]
                puts '予約情報に追加したよ'

                body_params[:seats] = body_params[:seats][0, body_params[:adult] + body_params[:child]]
                body_params[:car_number] = carnum

                break
              end
            end

            if body_params[:seats].length.zero?
              db.query('ROLLBACK')
              halt_with_error 404, 'あいまい座席予約ができませんでした。指定した席、もしくは1車両内に希望の席数をご用意できませんでした。'
            end
          end
        else
          # 座席情報のValidate
          body_params[:seats].each do |z|
            puts "XXXX #{z}"

            seat_list = begin
              db.xquery(
                'SELECT * FROM `seat_master` WHERE `train_class` = ? AND `car_number` = ? AND `seat_column` = ? AND `seat_row` = ? AND `seat_class` = ?',
                body_params[:train_class],
                body_params[:car_number],
                z[:column],
                z[:row],
                body_params[:seat_class],
              )
            rescue Mysql2::Error => e
              puts e.message
              db.query('ROLLBACK')
              halt_with_error 400, e.message
            end

            if seat_list.to_a.empty?
              db.query('ROLLBACK')
              halt_with_error 404, 'リクエストされた座席情報は存在しません。号車・喫煙席・座席クラスなど組み合わせを見直してください'
            end
          end
        end

        reservations = begin
          db.xquery(
            'SELECT * FROM `reservations` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ? FOR UPDATE',
            date.strftime('%Y/%m/%d'),
            body_params[:train_class],
            body_params[:train_name],
          )
        rescue Mysql2::Error => e
          puts e.message
          db.query('ROLLBACK')
          halt_with_error 500, '列車予約情報の取得に失敗しました'
        end

        reservations.each do |reservation|
          break if body_params[:seat_class] == 'non-reserved'

          # train_masterから列車情報を取得(上り・下りが分かる)
          tmas = begin
            db.xquery(
              'SELECT * FROM `train_master` WHERE `date` = ? AND `train_class` = ? AND `train_name` = ?',
              date.strftime('%Y/%m/%d'),
              body_params[:train_class],
              body_params[:train_name],
            ).first
          rescue Mysql2::Error => e
            puts e.message
            db.query('ROLLBACK')
            halt_with_error 500, '列車データの取得に失敗しました'
          end

          if tmas.nil?
            db.query('ROLLBACK')
            halt_with_error 404, '列車データがみつかりません'
          end

          # 予約情報の乗車区間の駅IDを求める

          # From
          reserved_from_station = begin
            db.xquery(
              'SELECT * FROM `station_master` WHERE `name` = ?',
              reservation[:departure],
            ).first
          rescue Mysql2::Error => e
            puts e.message
            db.query('ROLLBACK')
            halt_with_error 500, '予約情報に記載された列車の乗車駅データの取得に失敗しました'
          end

          if reserved_from_station.nil?
            db.query('ROLLBACK')
            halt_with_error 404, '予約情報に記載された列車の乗車駅データがみつかりません'
          end

          # To
          reserved_to_station = begin
            db.xquery(
              'SELECT * FROM `station_master` WHERE `name` = ?',
              reservation[:arrival],
            ).first
          rescue Mysql2::Error => e
            puts e.message
            db.query('ROLLBACK')
            halt_with_error 500, '予約情報に記載された列車の降車駅データの取得に失敗しました'
          end

          if reserved_to_station.nil?
            db.query('ROLLBACK')
            halt_with_error 404, '予約情報に記載された列車の降車駅データがみつかりません'
          end

          # 予約の区間重複判定
          secdup = false
          if tmas[:is_nobori]
            # 上り
            if to_station[:id] < reserved_to_station[:id] && from_station[:id] <= reserved_to_station[:id]
              # pass
            elsif to_station[:id] >= reserved_from_station[:id] && from_station > reserved_from_station[:id]
              # pass
            else
              secdup = true
            end
          else
            # 下り
            if from_station[:id] < reserved_from_station[:id] && to_station[:id] <= reserved_from_station[:id]
              # pass
            elsif from_station[:id] >= reserved_to_station[:id] && to_station[:id] > reserved_to_station[:id]
              # pass
            else
              secdup = true
            end
          end

          if secdup
            # 区間重複の場合は更に座席の重複をチェックする
            seat_reservations = begin
              db.xquery(
                'SELECT * FROM `seat_reservations` WHERE `reservation_id` = ? FOR UPDATE',
                reservation[:reservation_id],
              )
            rescue Mysql2::Error => e
              puts e.message
              db.query('ROLLBACK')
              halt_with_error 500, '座席予約情報の取得に失敗しました'
            end

            seat_reservations.each do |v|
              body_params[:seats].each do |seat|
                if v[:car_number] == body_params[:car_number] && v[:seat_row] == seat[:row] && v[:seat_column] == seat[:column]
                  db.query('ROLLBACK')
                  puts "Duplicated #{reservation}"
                  halt_with_error 400, 'リクエストに既に予約された席が含まれています'
                end
              end
            end
          end
        end

        # 3段階の予約前チェック終わり

        # 自由席は強制的にSeats情報をダミーにする（自由席なのに席指定予約は不可）
        if body_params[:seat_class] == 'non-reserved'
          body_params[:seats] = []
          body_params[:car_number] = 0

          (body_params[:adult] + body_params[:child]).times do
            body_params[:seats] << {
              row: 0,
              column: '',
            }
          end
        end

        # 運賃計算
        fare = begin
          case body_params[:seat_class]
          when 'premium', 'reserved', 'non-reserved'
            fare_calc(date, from_station[:id], to_station[:id], body_params[:train_class], body_params[:seat_class])
          else
            raise Error, 'リクエストされた座席クラスが不明です'
          end
        rescue Error, ErrorNoRows => e
          db.query('ROLLBACK')
          puts "fareCalc #{e.message}"
          halt_with_error 400, e.message
        end

        sum_fare = (body_params[:adult] * fare) + (body_params[:child] * fare) / 2
        puts 'SUMFARE'

        # userID取得。ログインしてないと怒られる。
        user, status, message = get_user

        if status != 200
          db.query('ROLLBACK')
          puts message
          halt_with_error status, message
        end

        # 予約ID発行と予約情報登録
        begin
          db.xquery(
            'INSERT INTO `reservations` (`user_id`, `date`, `train_class`, `train_name`, `departure`, `arrival`, `status`, `payment_id`, `adult`, `child`, `amount`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)',
            user[:id],
            date.strftime('%Y/%m/%d'),
            body_params[:train_class],
            body_params[:train_name],
            body_params[:departure],
            body_params[:arrival],
            'requesting',
            'a',
            body_params[:adult],
            body_params[:child],
            sum_fare,
          )
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 400, "予約の保存に失敗しました。 #{e.message}"
        end

        id = db.last_id # 予約ID
        if id.nil?
          db.query('ROLLBACK')
          halt_with_error 500, '予約IDの取得に失敗しました'
        end

        # 席の予約情報登録
        # reservationsレコード1に対してseat_reservationstが1以上登録される
        body_params[:seats].each do |v|
          db.xquery(
            'INSERT INTO `seat_reservations` (`reservation_id`, `car_number`, `seat_row`, `seat_column`) VALUES (?, ?, ?, ?)',
            id,
            body_params[:car_number],
            v[:row],
            v[:column],
          )
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '座席予約の登録に失敗しました'
        end
      rescue => e
        puts e.message
        db.query('ROLLBACK')
        halt_with_error 500, e.message
      end

      response = {
        reservation_id: id,
        amount: sum_fare,
        is_ok: true
      }

      db.query('COMMIT')

      content_type :json
      response.to_json
    end

    post '/api/train/reservation/commit' do
      db.query('BEGIN')

      begin
        # 予約IDで検索
        reservation = begin
          db.xquery(
            'SELECT * FROM `reservations` WHERE `reservation_id` = ?',
            body_params[:reservation_id],
          ).first
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '予約情報の取得に失敗しました'
        end

        if reservation.nil?
          db.query('ROLLBACK')
          halt_with_error 404, '予約情報がみつかりません'
        end

        # 支払い前のユーザチェック。本人以外のユーザの予約を支払ったりキャンセルできてはいけない。
        user, status, message = get_user

        if status != 200
          db.query('ROLLBACK')
          puts message
          halt_with_error status, message
        end

        if reservation[:user_id] != user[:id]
          db.query('ROLLBACK')
          halt_with_error 403, '他のユーザIDの支払いはできません'
        end

        # 予約情報の支払いステータス確認
        if reservation[:status] == 'done'
          db.query('ROLLBACK')
          halt_with_error 403, '既に支払いが完了している予約IDです'
        end

        # 決済する
        pay_info = {
          card_token: body_params[:card_token],
          reservation_id: body_params[:reservation_id],
          amount: reservation[:amount],
        }

        payment_api = ENV['PAYMENT_API'] || 'http://payment:5000'

        uri = URI.parse("#{payment_api}/payment")
        req = Net::HTTP::Post.new(uri)
        req.body = {
          payment_information: pay_info
        }.to_json
        req['Content-Type'] = 'application/json'

        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = uri.scheme == 'https'
        res = http.start { http.request(req) }

        # リクエスト失敗
        if res.code != '200'
          db.query('ROLLBACK')
          puts res.code
          halt_with_error 500, '決済に失敗しました。カードトークンや支払いIDが間違っている可能性があります'
        end

        # リクエスト取り出し
        output = begin
          JSON.parse(res.body, symbolize_names: true)
        rescue JSON::ParserError => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, 'JSON parseに失敗しました'
        end

        # 予約情報の更新
        begin
          db.xquery(
            'UPDATE `reservations` SET `status` = ?, `payment_id` = ? WHERE `reservation_id` = ?',
            'done',
            output[:payment_id],
            body_params[:reservation_id],
          )
        rescue Mysql2::Error => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, '予約情報の更新に失敗しました'
        end
      rescue => e
        puts e.message
        db.query('ROLLBACK')
        halt_with_error 500, e.message
      end

      rr = {
        is_ok: true
      }

      db.query('COMMIT')

      content_type :json
      rr.to_json
    end

    get '/api/auth' do
      user, status, message = get_user

      if status != 200
        puts message
        halt_with_error status, message
      end

      content_type :json
      { email: user[:email] }.to_json
    end

    post '/api/auth/signup' do
      salt = SecureRandom.random_bytes(1024)
      super_secure_password = OpenSSL::PKCS5.pbkdf2_hmac(
        body_params[:password],
        salt,
        100,
        256,
        'sha256',
      )

      db.xquery(
        'INSERT INTO `users` (`email`, `salt`, `super_secure_password`) VALUES (?, ?, ?)',
        body_params[:email],
        salt,
        super_secure_password,
      )

      message_response('registration complete')
    rescue Mysql2::Error => e
      puts e.message
      halt_with_error 502, 'user registration failed'
    end

    post '/api/auth/login' do
      user = db.xquery(
        'SELECT * FROM `users` WHERE `email` = ?',
        body_params[:email],
      ).first

      halt_with_error 403, 'authentication failed' if user.nil?

      challenge_password = OpenSSL::PKCS5.pbkdf2_hmac(
        body_params[:password],
        user[:salt],
        100,
        256,
        'sha256',
      )

      halt_with_error 403, 'authentication failed' if user[:super_secure_password] != challenge_password

      session[:user_id] = user[:id]

      message_response 'autheticated'
    end

    post '/api/auth/logout' do
      session[:user_id] = 0

      message_response 'logged out'
    end

    get '/api/user/reservations' do
      user, status, message = get_user

      if status != 200
        halt_with_error status, message
      end

      reservation_list = db.xquery(
        'SELECT * FROM `reservations` WHERE `user_id` = ?',
        user[:id],
      )

      reservation_response_list = reservation_list.to_a.map do |r|
        make_reservation_response(r)
      end

      content_type :json
      reservation_response_list.to_json
    end

    get '/api/user/reservations/:item_id' do
      user, status, message = get_user

      if status != 200
        halt_with_error status, message
      end

      item_id = params[:item_id].to_i
      if item_id <= 0
        halt_with_error 400, 'incorrect item id'
      end

      reservation = db.xquery(
        'SELECT * FROM `reservations` WHERE `reservation_id` = ? AND `user_id` = ?',
        item_id,
        user[:id],
      ).first

      halt_with_error 404, 'Reservation not found' if reservation.nil?

      reservation_response = make_reservation_response(reservation)

      content_type :json
      reservation_response.to_json
    end

    post '/api/user/reservations/:item_id/cancel' do
      user, code, message = get_user

      if code != 200
        halt_with_error code, message
      end

      item_id = params[:item_id].to_i
      if item_id <= 0
        halt_with_error 400, 'incorrect item id'
      end

      db.query('BEGIN')

      reservation = begin
        db.xquery(
          'SELECT * FROM `reservations` WHERE `reservation_id` = ? AND `user_id` = ?',
          item_id,
          user[:id],
        ).first
      rescue Mysql2::Error => e
        db.query('ROLLBACK')
        puts e.message
        halt_with_error 500, '予約情報の検索に失敗しました'
      end

      if reservation.nil?
        db.query('ROLLBACK')
        halt_with_error 404, 'reservations naiyo'
      end

      case reservation[:status]
      when 'rejected'
        db.query('ROLLBACK')
        halt_with_error 500, '何らかの理由により予約はRejected状態です'
      when 'done'
        # 支払いをキャンセルする
        payment_api = ENV['PAYMENT_API'] || 'http://payment:5000'

        uri = URI.parse("#{payment_api}/payment/#{reservation[:payment_id]}")
        req = Net::HTTP::Delete.new(uri)
        req.body = {
          payment_id: reservation[:payment_id]
        }.to_json
        req['Content-Type'] = 'application/json'

        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = uri.scheme == 'https'
        res = http.start { http.request(req) }

        # リクエスト失敗
        if res.code != '200'
          db.query('ROLLBACK')
          puts res.code
          halt_with_error 500, '決済に失敗しました。支払いIDが間違っている可能性があります'
        end

        # リクエスト取り出し
        output = begin
          JSON.parse(res.body, symbolize_names: true)
        rescue JSON::ParserError => e
          db.query('ROLLBACK')
          puts e.message
          halt_with_error 500, 'JSON parseに失敗しました'
        end

        puts output
      else
        # pass
      end

      begin
        db.xquery(
          'DELETE FROM `reservations` WHERE `reservation_id` = ? AND `user_id` = ?',
          item_id,
          user[:id],
        )
      rescue Mysql2::Error => e
        db.query('ROLLBACK')
        puts e.message
        halt_with_error 500, e.message
      end

      begin
        db.xquery(
          'DELETE FROM `seat_reservations` WHERE `reservation_id` = ?',
          item_id,
        )
      rescue Mysql2::Error => e
        db.query('ROLLBACK')
        puts e.message
        halt_with_error 500, e.message
      end

      db.query('COMMIT')

      message_response 'cancell complete'
    end

    error do |e|
      content_type :json
      { is_error: true, message: e.message }.to_json
    end
  end
end
