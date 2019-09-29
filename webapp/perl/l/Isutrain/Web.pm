package Isutrain::Web;

use strict;
use warnings;
use utf8;
use Kossy;

use JSON::XS 3.00;
use JSON::Types;
use DBIx::Sunny;
use Plack::Session;
use HTTP::Status qw/:constants/;

use Isutrain::Utils;

our $AVAILABLE_DAYS  = 10;
our $DEFAULT_PAYMENT_API  = "http://localhost:5000";

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my $host = $ENV{MYSQL_HOST} // '127.0.0.1';
        my $port = $ENV{MYSQL_PORT} // 3306;
        my $database = $ENV{MYSQL_DBNAME} // 'isucari';
        my $user = $ENV{MYSQL_USER} // 'isucari';
        my $password = $ENV{MYSQL_PASS} // 'isucari';
        my $dsn = "dbi:mysql:database=$database;host=$host;port=$port";
        DBIx::Sunny->connect($dsn, $user, $password, {
            mysql_enable_utf8mb4 => 1,
            mysql_auto_reconnect => 1,
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    # XXX $dbh->do('SET SESSION sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"');
                    return;
                },
            },
        });
    };
}

sub error_with_msg {
    my ($self, $c, $status, $msg) = @_;
    $c->res->code($status);
    $c->res->content_type('application/json;charset=utf-8');
    $c->res->body(JSON::encode_json({
        is_error => bool 1,
        message => $msg,
    }));
    $c->res;
}

sub fareCalc {
    my ($self, $date, $dep_station, $dest_station, $train_class, $seat_class) = @_;
    #
	# 料金計算メモ
	# 距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)
	#

    my $query = 'SELECT * FROM station_master WHERE id=?';
    my $from_station = $self->dbh->select_row(
        $query,
        $dep_station,
    );
    if (!$from_station) {
        die "no dep/from station";
    }

    my $to_station = $self->dbh->select_row(
        $query,
        $dest_station,
    );
    if (!$to_station) {
        die "no dest/to station";
    }

    warn("distance ", abs($to_station->{distance} - $from_station->{distance}));
    my $dist_fare = $self->getDistanceFare(abs($to_station->{distance} - $from_station->{distance}));
    warn("distFare", $dist_fare);

	# 期間・車両・座席クラス倍率
    my $fare_list = $self->dbh->select_all(
        'SELECT * FROM fare_master WHERE train_class=? AND seat_class=? ORDER BY start_date',
        $train_class,
        $seat_class,
    );
    if (@$fare_list == 0) {
        die "fare_master does not exists";
    }

	my $selected_fare = $fare_list->[0];

    for my $fare (@$fare_list) {
        my $start = $fare->{start_date};
        $start =~ s/ /T/;
        $start .= "+09:00";
        my $start_date = Time::Moment->from_string($start);
        if ($start_date < $date) {
            warn($fare->{start_date}, $fare->{fare_multiplier}); #XXX
            $selected_fare = $fare;
        }
    }
    warn('%%%%%%%%%%%%%%%%%%%'); #XXX
    return int(
        $dist_fare * $selected_fare->{fare_multiplier}
    );
}

sub getDistanceFare {
    my ($self, $orig_to_dest_distance) = @_;
    my $query = 'SELECT distance,fare FROM distance_fare_master ORDER BY distance';
    my $distance_fare_list = $self->dbh->select_all($query);

    my $last_distance = 0.0; #XXX
    my $last_fare = 0;

    for my $distance_fare (@$distance_fare_list) {
        warn($orig_to_dest_distance, $distance_fare->{distance}, $distance_fare->{fare});
        if ($last_distance < $orig_to_dest_distance && $orig_to_dest_distance < $distance_fare->{distance}) {
            last;
        }
        $last_distance = $distance_fare->{distance};
        $last_fare = $distance_fare->{fare};
    }
	return $last_fare;
}


#mux.HandleFunc(pat.Post("/initialize"), initializeHandler)
post '/initialize' => sub {
    my ($self, $c) = @_;

    $self->dbh->query("TRUNCATE seat_reservations");
    $self->dbh->query("TRUNCATE reservations");
    $self->dbh->query("TRUNCATE users");

    $c->render_json({
        available_days => $AVAILABLE_DAYS,
    });
};

#mux.HandleFunc(pat.Get("/api/settings"), settingsHandler)
get '/api/settings' => sub {
    my ($self, $c) = @_;

    my $payment_api = $ENV{PAYMENT_API} // $DEFAULT_PAYMENT_API;

    $c->render_json({
        payment_api => $payment_api,
    });
};

#mux.HandleFunc(pat.Get("/api/stations"), getStationsHandler)
get '/api/stations' => sub {
=comment
        駅一覧
            GET /api/stations

        return []Station{}
=cut
    my ($self, $c) = @_;

    my $sts = $self->dbh->select_all('SELECT * FROM station_master ORDER BY id');
    my @stations;
    for my $st (@$sts) {
        push @stations, {
            id => number $st->{id},
            name => string $st->{name},
            is_stop_express => bool $st->{is_stop_express},
            is_stop_semi_express => bool $st->{is_stop_semi_express},
            is_stop_local => bool $st->{is_stop_local},
        };
    }
    $c->render_json(\@stations);
};

#mux.HandleFunc(pat.Get("/api/train/search"), trainSearchHandler)
get '/api/train/search' => sub {
=comment
		列車検索
			GET /train/search?use_at=<ISO8601形式の時刻> & from=東京 & to=大阪

		return
			料金
			空席情報
			発駅と着駅の到着時刻

=cut
    my ($self, $c) = @_;

    my $jst_offset = 9*60;
    my $date = eval {
        Time::Moment->from_string($c->req->parameters->get('use_at'));
    };
    if ($@) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
    }
    $date = $date->with_offset_same_instant($jst_offset);

    if (!Isutrain::Utils::checkAvailableDate($AVAILABLE_DAYS, $date)) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "予約可能期間外です");
    }

    my $train_class =$c->req->parameters->get('train_class') // "";
    my $from_name =$c->req->parameters->get('from_name') // "";
    my $to_name =$c->req->parameters->get('to_name') // "";

    my $adult = number $c->req->parameters->get('adult');
    my $child = number $c->req->parameters->get('child');

    my $query = 'SELECT * FROM station_master WHERE name=?';
    # From
    my $from_station = $self->dbh->select_row(
        $query,
        $from_name
    );
    if (!$from_station) {
        warn("fromStation: no rows");
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "No From station");
    }

    # To
    my $to_station = $self->dbh->select_row(
        $query,
        $to_name
    );
    if (!$to_station) {
        warn("toStation: no rows");
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "No To station");
    }

    my $is_nobori = 0;
    if ( $from_station->{distance} > $to_station->{distance} ) {
        $is_nobori = 1;
    }

    $query = 'SELECT * FROM station_master ORDER BY distance';
    if ($is_nobori) {
        # 上りだったら駅リストを逆にする
        $query .= ' DESC';
    }

    my $usable_train_class_list = Isutrain::Utils::getUsableTrainClassList($from_station, $to_station);
    my $in_query;
    my @in_args = (
        $date->strftime("%F"),
        $usable_train_class_list,
        $is_nobori
    );
    if ($train_class == "") {
        $in_query = 'SELECT * FROM train_master WHERE date=? AND train_class IN (?) AND is_nobori=?'
    } else {
        $in_query = 'SELECT * FROM train_master WHERE date=? AND train_class IN (?) AND is_nobori=? AND train_class=?';
        push @in_args, $train_class;
    }

    my $tranin_list = $self->dbh->select_all($in_query, @in_args);
    my $stations = $self->dbh->select_all($query);

    warn("From ", $from_station); #XXX
    warn("To ", $to_station); #XXX

    my @train_search_response_list;

    for my $train (@$tranin_list) {
        my $is_seeked_to_first_station = 0;
        my $is_contains_origin_station = 0;
        my $is_contains_dest_station = 0;
        my $i = 0;

        for my $station (@$stations) {

            if (!$is_seeked_to_first_station) {
                # 駅リストを列車の発駅まで読み飛ばして頭出しをする
                # 列車の発駅以前は止まらないので無視して良い
                if ($station->{name} == $train->{start_station}) {
                    $is_seeked_to_first_station = 1;
                } else {
                    next;
                }
            }

            if ($station->{id} == $from_station->{id}) {
                # 発駅を経路中に持つ編成の場合フラグを立てる
                $is_contains_origin_station = 1;
            }
            if ($station->{id} == $to_station->{id}) {
                if ($is_contains_origin_station) {
                    # 発駅と着駅を経路中に持つ編成の場合
                    $is_contains_dest_station = 1;
                    last;
                } else {
                    # 出発駅より先に終点が見つかったとき
                    warn("なんかおかしい");
                    last;
                }
            }
            if ($station->{name} eq $train->{last_station}) {
                # 駅が見つからないまま当該編成の終点に着いてしまったとき
                last;
            }
            $i++;
        }

        if ($is_contains_origin_station && $is_contains_dest_station) {
            # 列車情報
            # 所要時間
            my $departure = $self->dbh->select_row(
                'SELECT departure FROM train_timetable_master WHERE date=? AND train_class=? AND train_name=? AND station=?',
                $date->strftime("%F"),
                $train->{train_class},
                $train->{train_name},
                $from_station->{name}
            );
            if (!$departure) {
                return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "No departure found");
            }

            my $departure_date = eval {
                Time::Moment->from_string(
                    sprintf("%sT%s+09:00",$date->strftime("%F"),$departure),
                );
            };
            if ($@) {
                return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, $@);
            }

            if (!($date < $departure_date)) {
                # 乗りたい時刻より出発時刻が前なので除外
                next;
            }

            my $arrival = $self->dbh->select_row(
                'SELECT arrival FROM train_timetable_master WHERE date=? AND train_class=? AND train_name=? AND station=?',
                $date->strftime("%F"),
                $train->{train_class},
                $train->{train_name},
                $to_station->{name}
            );
            if (!$arrival) {
                return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "No arrival found");
            }

            my $premium_avail_seats = Isutrain::Utils::getAvailableSeats(
                $self->dbh,
                $train,
                $from_station,
                $to_station,
                "premium",
                0
            );
            my $premium_smoke_avail_seats = Isutrain::Utils::getAvailableSeats(
                $self->dbh,
                $train,
                $from_station,
                $to_station,
                "premium",
                1
            );
            my $reserved_avail_seats = Isutrain::Utils::getAvailableSeats(
                $self->dbh,
                $train,
                $from_station,
                $to_station,
                "reserved",
                0
            );
            my $reserved_smoke_avail_seats = Isutrain::Utils::getAvailableSeats(
                $self->dbh,
                $train,
                $from_station,
                $to_station,
                "reserved",
                1
            );

            my $premium_avail = "○";
			if (@$premium_avail_seats == 0) {
				$premium_avail = "×";
			} elsif (@$premium_avail_seats < 10) {
				$premium_avail = "△";
			}

			my $premium_smoke_avail = "○";
			if (@$premium_smoke_avail_seats == 0) {
				$premium_smoke_avail = "×";
			} elsif (@$premium_smoke_avail_seats < 10) {
				$premium_smoke_avail = "△";
			}

			my $reserved_avail = "○";
			if (@$reserved_avail_seats == 0) {
				$reserved_avail = "×";
			} elsif (@$reserved_avail_seats < 10) {
				$reserved_avail = "△";
			}

			my $reserved_smoke_avail = "○";
			if (@$reserved_smoke_avail_seats == 0) {
				$reserved_smoke_avail = "×";
			} elsif (@$reserved_smoke_avail_seats < 10) {
				$reserved_smoke_avail = "△";
			}

            # TODO: 空席情報
			my %seat_availability = (
				"premium"        => $premium_avail,
				"premium_smoke"  => $premium_smoke_avail,
				"reserved"       => $reserved_avail,
				"reserved_smoke" => $reserved_smoke_avail,
				"non_reserved"   => "○",
            );

            # 料金計算
            my $premium_fare = eval {
                $self->fareCalc(
                    $date, $from_station->{id}, $to_station->{id},
                    $train->{train_class}, "premium"
                )
            };
            if ($@) {
                return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
            }
			$premium_fare = $premium_fare*$adult + $premium_fare/2*$child;

            my $reserved_fare = eval {
                $self->fareCalc(
                    $date, $from_station->{id}, $to_station->{id},
                    $train->{train_class}, "reserved"
                )
            };
            if ($@) {
                return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
            }
			$reserved_fare = $reserved_fare*$adult + $reserved_fare/2*$child;

            my $non_reserved_fare = eval {
                $self->fareCalc(
                    $date, $from_station->{id}, $to_station->{id},
                    $train->{train_class}, "non-reserved"
                )
            };
            if ($@) {
                return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
            }
			$non_reserved_fare = $non_reserved_fare*$adult + $non_reserved_fare/2*$child;

			my %fare_information = (
				"premium"        => number $premium_fare,
				"premium_smoke"  => number $premium_fare,
				"reserved"       => number $reserved_fare,
				"reserved_smoke" => number $reserved_fare,
				"non_reserved"   => number $non_reserved_fare,
			);

            push @train_search_response_list, {
                train_class => $train->{train_class},
                train_name => $train->{train_name},
                start => $train->{start_station},
                last => $train->{last_station},
                departure => $from_station->{name},
                arrival => $to_station->{name},
                departure_time => $departure,
                arrival_time => $arrival,
                seat_availability => \%seat_availability,
                seat_fare => \%fare_information,
            };

            if (@train_search_response_list >= 10) {
                last;
            }
		}
	}

    $c->render_json(\@train_search_response_list);
};


#mux.HandleFunc(pat.Get("/api/train/seats"), trainSeatsHandler)
#mux.HandleFunc(pat.Post("/api/train/reserve"), trainReservationHandler)
#mux.HandleFunc(pat.Post("/api/train/reservation/commit"), reservationPaymentHandler)

#mux.HandleFunc(pat.Get("/api/auth"), getAuthHandler)
#mux.HandleFunc(pat.Post("/api/auth/signup"), signUpHandler)
#mux.HandleFunc(pat.Post("/api/auth/login"), loginHandler)
#mux.HandleFunc(pat.Post("/api/auth/logout"), logoutHandler)
#mux.HandleFunc(pat.Get("/api/user/reservations"), userReservationsHandler)
#mux.HandleFunc(pat.Get("/api/user/reservations/:item_id"), userReservationResponseHandler)
#mux.HandleFunc(pat.Post("/api/user/reservations/:item_id/cancel"), userReservationCancelHandler)

1;
