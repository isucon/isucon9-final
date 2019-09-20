import random
import datetime

'''
    10分に1本 -> 1日192本
    192 * 366日 -> 70272レコード
'''

train_name = ['最速', '中間', '遅いやつ']
train_probability = [0.5, 0.25, 0.25]
src_dest = [
    ('東京', '大阪'),
    ('東京', '名古屋'),
    ('東京', '京都'),
    ('東京', '大阪'),
]

print('BEGIN;')
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

        print('INSERT INTO train_master(date,train_class,train_name,departure_at,start_station,last_station,is_nobori) VALUES ("%s","%s",%d,"%s","%s","%s","%d");' % (date.strftime("%Y-%m-%d"), name,i,t, dest[0],dest[1], 1 if is_nobori else 0))
    date = date + datetime.timedelta(days=1)
print('COMMIT;')
