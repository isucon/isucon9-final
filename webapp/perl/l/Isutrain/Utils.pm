package Isutrain::Utils;

use strict;
use warnings;
use utf8;
use Time::Moment;

our %TRAIN_CLASS_MAP = (
    "express" => "最速",
    "semi_express" => "中間",
    "local" => "遅いやつ"
);

sub checkAvailableDate {
    my ($available_days, $date) = @_;
    my $jst_offset = 9*60;
    my $t = Time::Moment->new(
        year  => 2020,
        month => 1,
        day   => 1,
        hour  => 0,
        minute => 0,
        second => 0,
        offset => $jst_offset,
    );
    $t = $t->plus_days($available_days);
    return $date < $t;
}

sub getUsableTrainClassList {
    my ($from_station, $to_station) = @_;

    my %usable;
    for my $key (keys %TRAIN_CLASS_MAP) {
        $usable{$key} = $TRAIN_CLASS_MAP{$key};
    }

    if (!$from_station->{is_stop_express}) {
        delete $usable{"express"};
    }
    if (!$from_station->{is_stop_semi_express}) {
        delete $usable{"semi_express"};
    }
    if (!$from_station->{is_stop_local}) {
        delete $usable{"local"}
    }

    if (!$to_station->{is_stop_express}) {
        delete $usable{"express"};
    }
    if (!$to_station->{is_stop_semi_express}) {
        delete $usable{"semi_express"};
    }
    if (!$to_station->{is_stop_local}) {
        delete $usable{"local"}
    }

    my @ret = values %usable;
    return \@ret;
}

sub getAvailableSeats {
    my ($dbh, $train, $from_station, $to_station, $seat_class, $is_smoking_seat) = @_;
    # 指定種別の空き座席を返す

    # 全ての座席を取得する
    my $query = 'SELECT * FROM seat_master WHERE train_class=? AND seat_class=? AND is_smoking_seat=?';
    my $seat_list = $dbh->select_all(
        $query,
        $train->{train_class},
        $seat_class,
        $is_smoking_seat
    );

    my %available_seat_map = ();
    for my $seat (@$seat_list) {
        my $key = sprintf("%d_%d_%s", $seat->{car_number}, $seat->{seat_row}, $seat->{seat_column});
        $available_seat_map{$key} = $seat;
    }

    # すでに取られている予約を取得する
    $query = <<'EOF';
    SELECT sr.reservation_id, sr.car_number, sr.seat_row, sr.seat_column
    FROM seat_reservations sr, reservations r, seat_master s, station_master std, station_master sta
    WHERE
        r.reservation_id=sr.reservation_id AND
        s.train_class=r.train_class AND
        s.car_number=sr.car_number AND
        s.seat_column=sr.seat_column AND
        s.seat_row=sr.seat_row AND
        std.name=r.departure AND
        sta.name=r.arrival
EOF

    if ($train->{is_nobori}) {
        $query .= "AND ((sta.id < ? AND ? <= std.id) OR (sta.id < ? AND ? <= std.id) OR (? < sta.id AND std.id < ?))"
    } else {
        $query .= "AND ((std.id <= ? AND ? < sta.id) OR (std.id <= ? AND ? < sta.id) OR (sta.id < ? AND ? < std.id))"
    }

    my $seat_reservation_list = $dbh->select_all(
        $query,
        $from_station->{id},
        $from_station->{id},
        $to_station->{id},
        $to_station->{id},
        $from_station->{id},
        $to_station->{id},
    );

    for my $seat_reservation (@$seat_reservation_list) {
        my $key = sprintf(
            "%d_%d_%s",
            $seat_reservation->{car_number},
            $seat_reservation->{seat_row},
            $seat_reservation->{seat_column},
        );
        delete $available_seat_map{$key};
    }

    my @ret = values %available_seat_map;
    return \@ret;
}
1;
