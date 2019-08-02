import random

'''
    席数のパターン
        先頭車両 x2
            5x13 ABC   DE
        通常車両 x5
            5x20 ABC   DE
        通常だけどトイレとかある車両 x6
            5x16 ABC   DE
            3, 5, 7, 11, 13, 15両目
        プレミアム(グリーン車) x3
            4x17 AB    CD

        計16両 1編成1314席
        1日252288席
        1年(366日)で92337408席
       
    自由席車両の数
        のぞみ的なやつ 3両
        ひかり的なやつ 5両
        こだま的なやつ 12もしくは13両(グリーン以外全部)
        指定席の設定がないので取れないパターンができる
        1両目から自由席で埋めていって残りが指定席
    グリーンは常に3両
        8,9,10両目

    
    | seat_master          | train_class           | varchar(100)                              | NO          |            | NULL           |                |
    | seat_master          | car_number            | int(11)                                   | NO          |            | NULL           |                |
    | seat_master          | seat_column           | enum('A','B','C','D','E')                 | NO          |            | NULL           |                |
    | seat_master          | seat_row              | int(11)                                   | NO          |            | NULL           |                |
    | seat_master          | seat_class            | enum('premium','reserved','non-reserved') | NO          |            | NULL           |                |
    | seat_master          | is_smoking_seat       | tinyint(1)                                | NO          |            | NULL           |                |
'''

train_name = ['最速', '中間', '遅いやつ']

print('BEGIN;')

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

                print('INSERT INTO seat_master(train_class,car_number,seat_column,seat_row,seat_class,is_smoking_seat) VALUES("%s",%d,"%s",%d,"%s",%d);' 
                    % (train_class, car_num, 'ABCDE'[column], row, seat_class, 1 if is_smoking_seat else 0))

print('COMMIT;')
