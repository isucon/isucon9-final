import random
import datetime
import csv

# 列車種別と出現割合
train_name = ['最速', '中間', '遅いやつ']
train_probability = [0.5, 0.25, 0.25]

# 列車種別ごとの料金倍率
train_fare_scale = [
    ('最速', 1.5),
    ('中間', 1.0),
    ('遅いやつ', 0.8),
]

# 座席種別ごとの料金倍率
seat_classes = [
    ('premium', 2.0),
    ('reserved', 1.25),
    ('non-reserved', 1.0),
]

# 期間運賃倍率
# 参考資料 https://www.ana.co.jp/ja/jp/amc/reference/tukau/award/dom/terms.html
season_map = [
    ("2020-01-01", 5.0), # 正月
    ("2020-01-06", 1.0),
    ("2020-03-13", 3.0), # 春休み
    ("2020-04-01", 1.0),
    ("2020-04-24", 5.0), # GW
    ("2020-05-11", 1.0),
    ("2020-08-07", 3.0), # 夏休み
    ("2020-08-24", 1.0),
    ("2020-12-25", 5.0), # 年越し
]

# 発駅-終点となる駅名のペア
# 逆向きも自動で作成される
src_dest = [
    ('東京', '大阪'),
    ('東京', '大阪'), # 多めにする
    ('東京', '名古屋'),
    ('東京', '京都'),
]

train_average_speed = [500, 480, 480] # km/h
station_stop_time = [1, 2, 2] # 分

train_data = []
station_data = []

def common_queries(file):
    queries = [
        'use isutrain;',
        'SET CHARACTER_SET_CLIENT = utf8;',
        'SET CHARACTER_SET_CONNECTION = utf8;']
    file.writelines('\n'.join(queries))
    file.write('\n\n')

def train_generator(filename):
    f = open(filename, 'w')
    common_queries(f)

    values = []
    f.write('INSERT INTO train_master(date,train_class,train_name,departure_at,start_station,last_station,is_nobori) VALUES\n\t')

    date = datetime.datetime(2020,1,1)
    for day in range(366):
        departure_time = [datetime.time(6, 0, 0), datetime.time(6, 0, 0)]
        for i in range(1, 193):
            is_nobori = False
            # name = train_name[2]
            t = departure_time[i % 2]
            departure_time[i % 2] = (datetime.datetime.combine(datetime.date(2019,8,1), departure_time[i % 2]) + datetime.timedelta(minutes=10+random.randint(-2,2))).time()
            if random.random() < train_probability[0]:
                name = train_name[0]
            elif random.random() < train_probability[1]:
                name = train_name[1]
            else:
                name = train_name[2]

            dest = random.choice(src_dest)
            if i % 2 == 0:
                dest = (dest[1], dest[0])
                is_nobori = True

            train_data.append((date.strftime("%Y-%m-%d"), name,i,t, dest[0],dest[1], 1 if is_nobori else 0))
            values.append('("%s","%s",%d,"%s","%s","%s","%d")' % (date.strftime("%Y-%m-%d"), name,i,t, dest[0],dest[1], 1 if is_nobori else 0))
        date = date + datetime.timedelta(days=1)

    f.write(',\n\t'.join(values))
    f.write(';\n')

    f.close()

def station_generator(filename):
    f = open(filename, 'w')
    soreppoi = open('soreppoi.csv', 'r')
    common_queries(f)

    # 駅名だけはそれっぽいデータを作ってCSVにしておく
    # http://g-chan.dip.jp/square/archives/2012/12/post_334.html
    values = []
    f.write('INSERT INTO station_master(name,distance,is_stop_express,is_stop_semi_express,is_stop_local) VALUES\n\t')
    reader = csv.reader(soreppoi)
    for row in reader:
        station_data.append((row[0], row[1], row[2], row[3], row[4]))
        values.append('("%s",%s,%s,%s,%s)' % (row[0], row[1], row[2], row[3], row[4]))
    f.write(',\n\t'.join(values))
    f.write(';\n')

    soreppoi.close()
    f.close()

def fare_generator(filename):
    f = open(filename, 'w')
    common_queries(f)

    values = []
    f.write('INSERT INTO fare_master(train_class,seat_class,start_date,fare_multiplier) VALUES\n\t')
    for season in season_map:
        for train_class in train_fare_scale:
            for seat_class in seat_classes:
                values.append('("%s","%s","%s",%.3f)' % (train_class[0], seat_class[0], season[0], season[1] * seat_class[1] * train_class[1]))
                # print()
    f.write(',\n\t'.join(values))
    f.write(';\n')

    f.close()

def seat_generator(filename):
    f = open(filename, 'w')
    common_queries(f)

    values = []
    f.write('INSERT INTO seat_master(train_class,car_number,seat_column,seat_row,seat_class,is_smoking_seat) VALUES\n\t')
    for train_class in train_name:
        for car_num in range(1, 17):
            max_seat_row = 0
            max_seat_column = 0

            # グリーン車以外は横に5座席
            max_seat_column = 5

            if car_num == 1 or car_num == 16:
                # 先頭車両なので13列
                max_seat_row = 13
                if car_num == 1:
                    # 1両目は常に自由席
                    seat_class = 'non-reserved'
                else:
                    if train_class == '遅いやつ':
                        # 何も無ければ普通は指定席になってる
                        seat_class = 'reserved'
                        # if random.random() < 0.1:
                        #     # まれに全車両自由席なことがある
                        #     seat_class = 'non-reserved'

            elif car_num >= 8 and car_num <= 10:
                # グリーン車
                seat_class = 'premium'
                max_seat_row = 17
                max_seat_column = 4
            elif car_num in [3, 5, 7, 11, 13, 15]:
                # トイレ車両
                max_seat_row = 16

                if car_num == 3:
                    # 3両目は常に自由席
                    seat_class = 'non-reserved'
                elif train_class == train_name[2]:
                    # 各停なら15両目までは常に自由席
                    seat_class = 'non-reserved'
                elif car_num == 5 and train_class == train_name[1]:
                    # 準急なら5両目まで自由席
                    seat_class = 'non-reserved'
                else:
                    # それ以外は指定席
                    seat_class = 'reserved'
                
            else:
                # トイレのない普通の車両
                max_seat_row = 20
                if car_num == 2:
                    # 2両目は常に自由席
                    seat_class = 'non-reserved'
                elif train_class == train_name[2]:
                    # 各停なら15両目までは常に自由席
                    seat_class = 'non-reserved'
                elif car_num == 4 and train_class == train_name[1]:
                    # 準急なら5両目まで自由席なので4両目は自由席
                    seat_class = 'non-reserved'
                else:
                    # それ以外は指定席
                    seat_class = 'reserved'

            for row in range(1, max_seat_row + 1):
                for column in range(0, max_seat_column):
                    # 基本的にヤニ席ではない
                    is_smoking_seat = False

                    if max_seat_row == 16 and row > 10:
                        is_smoking_seat = True
                    # print('%s,%d両目,%d%s席,%s,%s' % (train_class, car_num, row, 'ABCDE'[column], seat_class, is_smoking_seat))

                    values.append('("%s",%d,"%s",%d,"%s",%d)' 
                        % (train_class, car_num, 'ABCDE'[column], row, seat_class, 1 if is_smoking_seat else 0))
    
    f.write(',\n\t'.join(values))
    f.write(';\n')

    f.close()

def train_timetable_generator(filename):
    f = open(filename % 0, 'w')
    common_queries(f)

    values = []
    f.write('INSERT INTO train_timetable_master(date,train_class,train_name,station,arrival,departure) VALUES\n\t')
    i = 0
    for train in train_data:
        train_class_id = train_name.index(train[1])
        last_station_position = 0.0
        time_now = datetime.datetime.combine(datetime.date.today(),train[3])
        for station in station_data:
            if (train_class_id == 0 and station[2] == '1') or (train_class_id == 1 and station[3] == '1') or (train_class_id == 2 and station[4] == '1'):

                if float(station[1]) > last_station_position:
                    distance = float(station[1]) - last_station_position
                else:
                    distance = last_station_position - float(station[1])

                dt = distance / train_average_speed[train_class_id] * 3600 # sec
                arrival = time_now + datetime.timedelta(seconds=dt)
                departure = arrival + datetime.timedelta(minutes=station_stop_time[train_class_id] + random.random() * train_class_id)


                values.append('("%s","%s","%s","%s","%s","%s")' 
                        % (train[0], train[1], train[2], station[0], arrival.strftime("%H:%M:%S"), departure.strftime("%H:%M:%S")))
                
                if len(values) > 500000:
                    f.write(',\n\t'.join(values))
                    f.write(';\n')
                    f.close()

                    i = i + 1

                    f = open(filename % i, 'w')
                    common_queries(f)
                    print("%d" % (500000 * (i)))

                    values = []
                    f.write('INSERT INTO train_timetable_master(date,train_class,train_name,station,arrival,departure) VALUES\n\t')

                time_now = departure
                last_station_position = float(station[1])
            
    f.write(',\n\t'.join(values))
    f.write(';\n')

    f.close()

if __name__ == '__main__':
    print('90_train.sql generating...', end='', flush=True)
    train_generator('90_train.sql')
    print('ok')

    print('91_station.sql generating...', end='', flush=True)
    station_generator('91_station.sql')
    print('ok')

    print('92_fare.sql generating...', end='', flush=True)
    fare_generator('92_fare.sql')
    print('ok')

    print('93_seat.sql generating...', end='', flush=True)
    seat_generator('93_seat.sql')
    print('ok')

    print('94_train_timetable.sql generating...', end='', flush=True)
    train_timetable_generator('94_%d_train_timetable.sql')
    print('ok')
