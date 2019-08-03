import random
import datetime

train_probability = [0.5, 0.25, 0.25]

train_fare_scale = [
    ('最速', 1.5),
    ('中間', 1.0),
    ('遅いやつ', 0.8),
]
seat_classes = [
    ('premium', 2.0),
    ('reserved', 1.25),
    ('non-reserved', 1.0),
]
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

for season in season_map:
    for train_class in train_fare_scale:
        for seat_class in seat_classes:
            print('INSERT INTO fare_master(train_class,seat_class,start_date,fare_multiplier) VALUES ("%s","%s","%s",%.3f);' % (train_class[0], seat_class[0], season[0], season[1] * seat_class[1] * train_class[1]))
