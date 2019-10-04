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
use Crypt::PBKDF2;
use Digest::SHA;
use Crypt::OpenSSL::Random;
use LWP::UserAgent;
use Isutrain::Utils;
use Log::Minimal;
use Time::Moment;


our $AVAILABLE_DAYS  = 10;
our $DEFAULT_PAYMENT_API  = "http://localhost:5000";

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my $host = $ENV{MYSQL_HOSTNAME} // '127.0.0.1';
        my $port = $ENV{MYSQL_PORT} // 3306;
        my $database = $ENV{MYSQL_DBNAME} // 'isutrain';
        my $user = $ENV{MYSQL_USER} // 'isutrain';
        my $password = $ENV{MYSQL_PASS} // 'isutrain';
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

sub secure_random_str {
    my $length = shift || 16;
    unpack("H*",Crypt::OpenSSL::Random::random_bytes($length))
}

sub getUser {
    my ($self, $c) = @_;
    my $session = Plack::Session->new($c->env);
    my $user_id = $session->get('user_id');
    return unless $user_id;
    return $self->dbh->select_row('SELECT * FROM users WHERE id = ?', $user_id);
}

sub error_with_msg {
    my ($self, $c, $status, $msg) = @_;
    $c->res->code($status);
    $c->res->content_type('application/json;charset=utf-8');
    $c->res->body(JSON::encode_json({
        is_error => JSON::true,
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
    warn("distFare ", $dist_fare);

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
            warnf("%s %s", $fare->{start_date}, $fare->{fare_multiplier});
            $selected_fare = $fare;
        }
    }
    warn('%%%%%%%%%%%%%%%%%%%');
    return int(
        $dist_fare * $selected_fare->{fare_multiplier}
    );
}

sub getDistanceFare {
    my ($self, $orig_to_dest_distance) = @_;
    my $query = 'SELECT distance,fare FROM distance_fare_master ORDER BY distance';
    my $distance_fare_list = $self->dbh->select_all($query);

    my $last_distance = 0.0;
    my $last_fare = 0;

    for my $distance_fare (@$distance_fare_list) {
        # warnf("%s %s %s", $orig_to_dest_distance, $distance_fare->{distance}, $distance_fare->{fare});
        if ($last_distance < $orig_to_dest_distance && $orig_to_dest_distance < $distance_fare->{distance}) {
            last;
        }
        $last_distance = $distance_fare->{distance};
        $last_fare = $distance_fare->{fare};
    }
	return $last_fare;
}

filter 'allow_json_request' => sub {
    my $app = shift;
    return sub {
        my ($self, $c) = @_;
        my $p = decode_json($c->req->content);
        my @params;
        if (ref $p eq 'HASH') {
            while (my ($k, $v) = each %$p) {
                push @params, $k, $v;
            }
        }
        my $hmv = Hash::MultiValue->new(@params);
        $c->env->{'kossy.request.body'} = $hmv;
        $c->env->{'plack.request.body'} = $hmv;
        $app->($self, $c);
    };
};

#mux.HandleFunc(pat.Post("/initialize"), initializeHandler)
post '/initialize' => sub {
    my ($self, $c) = @_;

    $self->dbh->query("TRUNCATE seat_reservations");
    $self->dbh->query("TRUNCATE reservations");
    $self->dbh->query("TRUNCATE users");

    $c->render_json({
        available_days => $AVAILABLE_DAYS,
        language => "perl",
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
    my $from_name =$c->req->parameters->get('from') // "";
    my $to_name =$c->req->parameters->get('to') // "";

    my $adult = number $c->req->parameters->get('adult') // 0;
    my $child = number $c->req->parameters->get('child') // 0;

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
    if ($train_class eq "") {
        $in_query = 'SELECT * FROM train_master WHERE date=? AND train_class IN (?) AND is_nobori=?'
    } else {
        $in_query = 'SELECT * FROM train_master WHERE date=? AND train_class IN (?) AND is_nobori=? AND train_class=?';
        push @in_args, $train_class;
    }

    my $tranin_list = $self->dbh->select_all($in_query, @in_args);
    my $stations = $self->dbh->select_all($query);

    warnf("From %s", $from_station);
    warnf("To %s", $to_station);

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
                if ($station->{name} eq $train->{start_station}) {
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
                    warnf("なんかおかしい");
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
            my $departure = $self->dbh->select_one(
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

            my $arrival = $self->dbh->select_one(
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
get '/api/train/seats' => sub {
    #
    # 指定した列車の座席列挙
    # GET /train/seats?date=2020-03-01&train_class=のぞみ&train_name=96号&car_number=2&from=大阪&to=東京
    #
    my ($self, $c) = @_;

    my $jst_offset = 9*60;
    my $date = eval {
        Time::Moment->from_string($c->req->parameters->get('date'));
    };
    if ($@) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
    }
    $date = $date->with_offset_same_instant($jst_offset);

    if (!Isutrain::Utils::checkAvailableDate($AVAILABLE_DAYS, $date)) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "予約可能期間外です");
    }

    my $train_class = $c->req->parameters->get('train_class') // "";
    my $train_name = $c->req->parameters->get('train_name') // "";
    my $car_number = $c->req->parameters->get('car_number') // "";
    my $from_name = $c->req->parameters->get('from') // "";
    my $to_name = $c->req->parameters->get('to') // "";

    # 対象列車の取得
    my $query = 'SELECT * FROM train_master WHERE date=? AND train_class=? AND train_name=?';
	my $train = $self->dbh->select_row($query, $date->strftime("%F"), $train_class, $train_name);
    if (!$train) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "列車が存在しません");
    }

    $query = 'SELECT * FROM station_master WHERE name=?';

	# From
    my $from_station = $self->dbh->select_row($query, $from_name);
    if (!$from_name) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "No From station");
    }

	# To
    my $to_station = $self->dbh->select_row($query, $to_name);
    if (!$to_station) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "No To station");
    }


    my $usable_train_class_list = Isutrain::Utils::getUsableTrainClassList($from_station, $to_station);
    my $usable = 0;
    for my $v (@$usable_train_class_list) {
        if ($v eq $train->{train_class}) {
            $usable = 1;
        }
    }
	if (!$usable) {
		warn("invalid train_class");
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "invalid train_class");
	}


    $query = 'SELECT * FROM seat_master WHERE train_class=? AND car_number=? ORDER BY seat_row, seat_column';
    my $seat_list = $self->dbh->select_all($query, $train_class, $car_number);

    my @seat_information_list = ();
    for my $seat (@$seat_list) {
        my $s = {
            row => number $seat->{seat_row},
            column => $seat->{seat_column},
            class => $seat->{seat_class},
            is_smoking_seat => bool $seat->{is_smoking_seat},
            is_occupied => JSON::false,
        };

        $query = <<'EOF';
        SELECT s.*
        FROM seat_reservations s, reservations r
        WHERE
        	r.date=? AND r.train_class=? AND r.train_name=? AND car_number=? AND seat_row=? AND seat_column=?
EOF
        my $seat_reservation_list = $self->dbh->select_all(
            $query,
            $date->strftime("%F"),
            $seat->{train_class},
            $train_name,
            $seat->{car_number},
            $seat->{seat_row},
            $seat->{seat_column}
        );

        warnf($seat_reservation_list);

        for my $seat_reservation (@$seat_reservation_list) {
            $query = 'SELECT * FROM reservations WHERE reservation_id=?';
            my $reservation = $self->dbh->select_row($query, $seat_reservation->{reservation_id});
            if (!$reservation) {
                die "No reservation";
            }

            $query = 'SELECT * FROM station_master WHERE name=?';
            my $departure_station = $self->dbh->select_row($query, $seat_reservation->{departure});
            if (!$departure_station) {
                die "No departure station";
            }
            my $arrival_station = $self->dbh->select_row($query, $seat_reservation->{arrival});
            if (!$arrival_station) {
                die "No arrival station";
            }

            if ($train->{is_nobori}) {
                # 上り
                if ($to_station->{id} < $arrival_station->{id}  && $from_station->{id} <= $arrival_station->{id}) {
                    # pass
                } elsif ($to_station->{id} >= $departure_station->{id} && $from_station->{id} > $departure_station->{id}) {
                    # pass
                } else {
                    $s->{is_occupied} = 1;
                }
            }
            else {
                # 下り
				if ($from_station->{id} < $departure_station->{id} && $to_station->{id} <= $departure_station->{id}) {
					# pass
				} elsif ($from_station->{id} >= $arrival_station->{id} && $to_station->{id} > $arrival_station->{id}) {
					# pass
				} else {
                    $s->{is_occupied} = 1;
				}
            }
        }

        warn($s->{is_occupied});
        push @seat_information_list, $s;
	}

	# 各号車の情報
    my @simple_car_information_list = ();

    $query = 'SELECT * FROM seat_master WHERE train_class=? AND car_number=? ORDER BY seat_row, seat_column LIMIT 1';
    my $i = 1;
    while (1) {
        my $seat = $self->dbh->select_row($query, $train_class, $i);
        if (!$seat) {
            last;
        }
        push @simple_car_information_list, {
            car_number => number $i,
            seat_class => $seat->{seat_class}
        };
        $i++;
    }

    my $date_str = $date->strftime("%F");
    $date_str =~ s/-/\//g;
    my $ci = {
        date => $date_str,
        train_class => $train_class,
        train_name => $train_name,
        car_number => number $car_number,
        seats => \@seat_information_list,
        cars => \@simple_car_information_list,
    };

    $c->render_json($ci);
};

#mux.HandleFunc(pat.Post("/api/train/reserve"), trainReservationHandler)
post '/api/train/reserve' => [qw/allow_json_request/] => sub {
=comment
        列車の席予約API　支払いはまだ
        POST /api/train/reserve
            {
                "date": "2020-12-31T07:57:00+09:00",
                "train_name": "183",
                "train_class": "中間",
                "car_number": 7,
                "is_smoking_seat": false,
                "seat_class": "reserved",
                "departure": "東京",
                "arrival": "名古屋",
                "child": 2,
                "adult": 1,
                "column": "A",
                "seats": [
                    {
                    "row": 3,
                    "column": "B"
                    },
                        {
                    "row": 4,
                    "column": "C"
                    }
                ]
        }
        レスポンスで予約IDを返す
        reservationResponse(w http.ResponseWriter, errCode int, id int, ok bool, message string)
=cut
    my ($self, $c) = @_;

    # 乗車日の日付表記統一
    my $jst_offset = 9*60;
    my $date = eval {
        Time::Moment->from_string($c->req->parameters->get('date'));
    };
    if ($@) {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "時刻のparseに失敗しました");
    }
    $date = $date->with_offset_same_instant($jst_offset);

    if (!Isutrain::Utils::checkAvailableDate($AVAILABLE_DAYS, $date)) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "予約可能期間外です");
    }

    my $train_class = $c->req->parameters->get('train_class') // "";
    my $train_name = $c->req->parameters->get('train_name') // "";
    my $departure = $c->req->parameters->get('departure') // "";
    my $arrival = $c->req->parameters->get('arrival') // "";
    my $seats = $c->req->parameters->get('seats') // [];
    my $seat_class = $c->req->parameters->get('seat_class') // "";
    my $car_number = $c->req->parameters->get('car_number') // 0;
    my $is_smoking_seat = $c->req->parameters->get('is_smoking_seat') // JSON::false;
    my $req_adult = $c->req->parameters->get('adult') // 0;
    my $req_child = $c->req->parameters->get('child') // 0;
    my $req_column = $c->req->parameters->get('column') // "";

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    # 止まらない駅の予約を取ろうとしていないかチェックする
	# 列車データを取得
    my $query = 'SELECT * FROM train_master WHERE date=? AND train_class=? AND train_name=?';
    my $tmas = $dbh->select_row(
        $query,
        $date->strftime("%F"),
        $train_class,
        $train_name,
    );
    if (!$tmas) {
        warn("no train");
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "列車データがみつかりません");
    }

	# 列車自体の駅IDを求める
	$query = 'SELECT * FROM station_master WHERE name=?';
    # Departure
    my $departure_station = $dbh->select_row($query, $tmas->{start_station});
    if (!$departure_station) {
        warn("no departure station");
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "リクエストされた列車の始発駅データがみつかりません");
    }

	# Arrive
    my $arrival_station = $dbh->select_row($query, $tmas->{last_station});
    if (!$arrival_station) {
        warn("no arrival station");
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "リクエストされた列車の終着駅データがみつかりません");
    }

	# リクエストされた乗車区間の駅IDを求める
	$query = 'SELECT * FROM station_master WHERE name=?';
	# From
    my $from_station = $dbh->select_row($query, $departure);
    if (!$from_station) {
        warn("no From station");
        return $self->error_with_msg($c, HTTP_NOT_FOUND,
            sprintf("乗車駅データがみつかりません %s", $departure)
        );
    }
	# To
    my $to_station = $dbh->select_row($query, $arrival);
    if (!$to_station) {
        warn("no To station");
        return $self->error_with_msg($c, HTTP_NOT_FOUND,
            sprintf("降車駅データがみつかりません %s", $arrival)
        );
    }

    if ($train_class eq "最速") {
        if (!$from_station->{is_stop_express} || !$to_station->{is_stop_express}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "最速の止まらない駅です");
        }
    } elsif ($train_class eq "中間") {
        if (!$from_station->{is_stop_semi_express} || !$to_station->{is_stop_semi_express}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "中間の止まらない駅です");
        }
    } elsif ($train_class eq "遅いやつ") {
        if (!$from_station->{is_stop_local} || !$to_station->{is_stop_local}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "遅いやつの止まらない駅です");
        }
    } else {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた列車クラスが不明です");
    }

    # 運行していない区間を予約していないかチェックする
    if ($tmas->{is_nobori}) {
        if ($from_station->{id} > $departure_station->{id} || $to_station->{id} > $departure_station->{id}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた区間に列車が運行していない区間が含まれています");
        }
        if ($arrival_station->{id} >= $from_station->{id} || $arrival_station->{id} > $to_station->{id}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた区間に列車が運行していない区間が含まれています");
        }
    }
    else {
        if ($from_station->{id} < $departure_station->{id} || $to_station->{id} < $departure_station->{id}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた区間に列車が運行していない区間が含まれています");

        }
        if ($arrival_station->{id} <= $from_station->{id} || $arrival_station->{id} < $to_station->{id}) {
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた区間に列車が運行していない区間が含まれています");
        }
    }

    #
    # あいまい座席検索
    # seatsが空白の時に発動する
    #

    if (@$seats == 0) {
        if ($seat_class eq "non-reserved") {
            # non-reservedはそもそもあいまい検索もせずダミーのRow/Columnで予約を確定させる。
            last;
        }
        # 当該列車・号車中の空き座席検索
        $query = 'SELECT * FROM train_master WHERE date=? AND train_class=? AND train_name=?';
        my $train = $dbh->select_row(
            $query,
            $date->strftime("%F"),
            $train_class,
            $train_name
        );
        if (!$train) {
            die "No train";
        }

        my $usable_train_class_list = Isutrain::Utils::getUsableTrainClassList($from_station, $to_station);
        my $usable = 0;
        for my $v (@$usable_train_class_list) {
            if ($v eq $train->{train_class}) {
                $usable = 1;
            }
        }
        if (!$usable) {
            warn("invalid train_class");
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "invalid train_class");

        }

        # 座席リクエスト情報は空に
        $seats = [];
        for (my $carnum = 1; $carnum <= 16; $carnum++) {
            $query = 'SELECT * FROM seat_master WHERE train_class=? AND car_number=? AND seat_class=? AND is_smoking_seat=? ORDER BY seat_row, seat_column';
            my $seat_list = $dbh->select_all($query, $train_class, $carnum, $seat_class, $is_smoking_seat);

            my @seat_information_list = ();
            for my $seat (@$seat_list) {
                my $s = {
                    row => number $seat->{seat_row},
                    column => $seat->{seat_column},
                    class => $seat->{seat_class},
                    is_smoking_seat => bool $seat->{is_smoking_seat},
                    is_occupied => JSON::false,
                };
                $query = 'SELECT s.* FROM seat_reservations s, reservations r WHERE r.date=? AND r.train_class=? AND r.train_name=? AND car_number=? AND seat_row=? AND seat_column=? FOR UPDATE';
                my $seat_reservation_list = $dbh->select_all(
                    $query,
                    $date->strftime("%F"),
                    $seat->{train_class},
                    $train_name,
                    $seat->{car_number},
                    $seat->{seat_row},
                    $seat->{seat_column}
                );

                for my $seat_reservation (@$seat_reservation_list) {
                    $query = 'SELECT * FROM reservations WHERE reservation_id=? FOR UPDATE';
                    my $reservation = $dbh->select_row($query, $seat_reservation->{reservation_id});
                    if (!$reservation) {
                        die "no reservations";
                    }

                    $query = 'SELECT * FROM station_master WHERE name=?';
                    my $departure_station = $dbh->select_row($query, $reservation->{departure});
                    if (!$departure_station) {
                        die "no departure";
                    }
                    my $arrival_station = $dbh->select_row($query, $reservation->{arrival});
                    if (!$arrival_station) {
                        die "no arrival";
                    }


                    if ($train->{is_nobori}) {
                        # 上り
                        if ($to_station->{id} < $arrival_station->{id} && $from_station->{id} <= $arrival_station->{id}) {
                            # pass
                        } elsif ($to_station->{id} >= $departure_station->{id} && $from_station->{id} > $departure_station->{id}) {
                            #pass
                        } else {
                            $s->{is_occupied} = JSON::true;
                        }
                    }
                    else {
                        # 下り
                        if ($from_station->{id} < $departure_station->{id} && $to_station->{id} <= $departure_station->{id}) {
                            # pass
                        } elsif ($from_station->{id} >= $arrival_station->{id} && $to_station->{id} > $arrival_station->{id}) {
                            # pass
                        } else {
                            $s->{is_occupied} = JSON::true;
                        }

                    }
                }
                push @seat_information_list, $s;
            }

            my $vague_seat = {};
            my $reserved = JSON::false;
            my $vargue = JSON::true;
            my $seatnum = ($req_adult + $req_child - 1); # 全体の人数からあいまい指定席分を引いておく
            if ($req_column eq "") {          # A/B/C/D/Eを指定しなければ、空いている適当な指定席を取るあいまいモード
                $seatnum = ($req_adult + $req_child); # あいまい指定せず大人＋小人分の座席を取る
                $reserved = JSON::true;                   # dummy
                $vargue = JSON::false;                    # dummy
            }
            my @candidate_seats;
            # シート分だけ回して予約できる席を検索
            my $i = 0;
            for my $seat (@seat_information_list) {
                if ($seat->{column} eq $req_column && !$seat->{is_occupied} && !$reserved && $vargue) {
                    #あいまい席があいてる
                    $vague_seat->{row} = number $seat->{row};
                    $vague_seat->{column} = $seat->{column};
                    $reserved = JSON::true
                } elsif (!$seat->{is_occupied} && $i < $seatnum) {
                    # 単に席があいてる
                    push @candidate_seats, {
                        row => number $seat->{row},
                        column => $seat->{column}
                    };
                    $i++;
                }
            }

            if ($vargue && $reserved) { # あいまい席が見つかり、予約できそうだった
                push @$seats, $vague_seat;
            }
            if ($i > 0) {
                push @$seats, @candidate_seats;
            }

            if (@$seats < $req_adult + $req_child) {
                # リクエストに対して席数が足りてない
                # 次の号車にうつしたい
                warn("-----------------");
                warnf(
                    "現在検索中の車両: %d号車, リクエスト座席数: %d, 予約できそうな座席数: %d, 不足数: %d",
                    $carnum,
                    $req_adult+$req_child,
                    scalar @$seats,
                    $req_adult+$req_child-scalar(@$seats)
                );
                warnf("リクエストに対して座席数が不足しているため、次の車両を検索します。");
                $seats = [];
                if ($carnum == 16) {
                    warnf("この新幹線にまとめて予約できる席数がなかったから検索をやめるよ");
                    $seats = [];
                    last;
                }
            }

            warnf(
                "空き実績: %d号車 シート:%s 席数:%d",
                $carnum,
                $seats,
                scalar @$seats
            );
            if (scalar @$seats >= $req_adult+$req_child) {
                warn("予約情報に追加したよ");
                while(@$seats > $req_adult+$req_child) { pop @$seats }
                $car_number = $carnum;
                last;
            }
        }

        if (@$seats == 0) {
            warnf("could not reserve aimai seat");
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "あいまい座席予約ができませんでした。指定した席、もしくは1車両内に希望の席数をご用意できませんでした。");
        }
    }
    else {
		# 座席情報のValidate
        for my $z (@$seats) {
            warnf("XXXX %s", $z);
            my $query = 'SELECT * FROM seat_master WHERE train_class=? AND car_number=? AND seat_column=? AND seat_row=? AND seat_class=?';
            my $seat_list = $dbh->select_row(
                $query,
                $train_class,
                $car_number,
                $z->{column},
                $z->{row},
                $seat_class,
            );
            if (!$seat_list) {
                warn("Not seat");
                return $self->error_with_msg($c, HTTP_NOT_FOUND, "あいまい座席予約ができませんでした。指定した席、もしくは1車両内に希望の席数をご用意できませんでした。");
            }
        }
	}

    # 当該列車・列車名の予約一覧取得
    $query = 'SELECT * FROM reservations WHERE date=? AND train_class=? AND train_name=? FOR UPDATE';
    my $reservations = $dbh->select_all(
        $query,
        $date->strftime("%F"),
        $train_class,
        $train_name
    );

    for my $reservation (@$reservations) {
        if ($seat_class eq "non-reserved") {
            last;
        }
        # train_masterから列車情報を取得(上り・下りが分かる)
        $query = 'SELECT * FROM train_master WHERE date=? AND train_class=? AND train_name=?';
        my $tmas = $dbh->select_row($query, $date->strftime("%F"), $train_class, $train_name);
        if (!$tmas) {
            warn("no train");
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "列車データがみつかりません");
        }

        # 予約情報の乗車区間の駅IDを求める
        $query = 'SELECT * FROM station_master WHERE name=?';
        # From
        my $reserved_from_station = $dbh->select_row($query, $reservation->{departure});
        if (!$reserved_from_station) {
            warn("no reserved from station");
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "予約情報に記載された列車の乗車駅データがみつかりません");
        }
        # To
        my $reserved_to_station = $dbh->select_row($query, $reservation->{arrival});
        if (!$reserved_to_station) {
            warn("no reserved to station");
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "予約情報に記載された列車の降車駅データがみつかりません");
        }

        # 予約の区間重複判定
        my $secdup = JSON::false;
        if ($tmas->{is_nobori}) {
            # 上り
            if ($to_station->{id} < $reserved_to_station->{id} && $from_station->{id} <= $reserved_to_station->{id}) {
                # pass
            } elsif ($to_station->{id} >= $reserved_from_station->{id} && $from_station->{id} > $reserved_from_station->{id}) {
                # pass
            } else {
                $secdup = JSON::true;
            }
        } else {
            # 下り
            if ($from_station->{id} < $reserved_from_station->{id} && $to_station->{id} <= $reserved_from_station->{id}) {
                # pass
            } elsif ($from_station->{id} >= $reserved_to_station->{id} && $to_station->{id} > $reserved_to_station->{id}) {
                # pass
            } else {
                $secdup = JSON::true;
            }
        }

        if ($secdup) {
            # 区間重複の場合は更に座席の重複をチェックする
            $query = 'SELECT * FROM seat_reservations WHERE reservation_id=? FOR UPDATE';
            my $seat_reservations = $dbh->select_all($query, $reservation->{reservation_id});
            for my $v (@$seat_reservations) {
                for my $seat (@$seats) {
                    if ($v->{car_number} == $car_number &&
                        $v->{seat_row}  == $seat->{row} &&
                        $v->{seat_column} eq $seat->{column}) {
                            warnf("Duplicated %s", $reservation);
                            return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストに既に予約された席が含まれています");
                    }
                }
            }
        }
	}
	# 3段階の予約前チェック終わり

	# 自由席は強制的にSeats情報をダミーにする（自由席なのに席指定予約は不可）
    if ($seat_class eq "non-reserved") {
        $seats = [];
        my $dummy_seat = {};
        $car_number = 0;
        for (my $num = 0; $num < $req_adult+$req_child; $num++) {
            $dummy_seat->{row} = 0;
            $dummy_seat->{column} = "";
            push @$seats, $dummy_seat;
        }
    }

    # 運賃計算
    my $fare = 0;
    if ($seat_class eq "premium") {
        $fare = eval {
            $self->fareCalc(
                $date, $from_station->{id}, $to_station->{id},
                $train_class, "premium"
            )
        };
        if ($@) {
            warn("fareCalc ", $@);
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
        }
    }
    elsif ($seat_class eq "reserved") {
        $fare = eval {
            $self->fareCalc(
                $date, $from_station->{id}, $to_station->{id},
                $train_class, "reserved"
            )
        };
        if ($@) {
            warn("fareCalc ", $@);
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
        }
    }
    elsif ($seat_class eq "non-reserved") {
        $fare = eval {
            $self->fareCalc(
                $date, $from_station->{id}, $to_station->{id},
                $train_class, "non-reserved"
            )
        };
        if ($@) {
            warn("fareCalc ", $@);
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
        }
    }
    else {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "リクエストされた座席クラスが不明です");
    }

    my $sum_fare = ($req_adult * $fare) + ($req_child * $fare)/2;
    warn("SUMFARE");

    # userID取得。ログインしてないと怒られる。
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    # 予約ID発行と予約情報登録
    $query = 'INSERT INTO `reservations` (`user_id`, `date`, `train_class`, `train_name`, `departure`, `arrival`, `status`, `payment_id`, `adult`, `child`, `amount`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
    $dbh->query(
        $query,
        $user->{id},
        $date->strftime("%F"),
        $train_class,
        $train_name,
        $departure,
        $arrival,
        "requesting",
        "a",
        $req_adult,
        $req_child,
        $sum_fare
    );

    my $id = $dbh->last_insert_id();
    if (!$id) {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, '予約IDの取得に失敗しました');
    }

    # 席の予約情報登録
    # reservationsレコード1に対してseat_reservationstが1以上登録される
    $query = 'INSERT INTO `seat_reservations` (`reservation_id`, `car_number`, `seat_row`, `seat_column`) VALUES (?, ?, ?, ?)';

    for my $v (@$seats) {
        $dbh->query(
            $query,
            $id,
            $car_number,
            $v->{row},
            $v->{column}
        );
    }

    $txn->commit();

    my $res = {
        reservation_id => number $id,
        amount => number $sum_fare,
        is_ok => JSON::true,
    };

    $c->render_json($res);
};

#mux.HandleFunc(pat.Post("/api/train/reservation/commit"), reservationPaymentHandler)
post '/api/train/reservation/commit' => [qw/allow_json_request/] => sub {
=comment
    支払い及び予約確定API
    POST /api/train/reservation/commit
    {
        "card_token": "161b2f8f-791b-4798-42a5-ca95339b852b",
        "reservation_id": "1"
    }

    前段でフロントがクレカ非保持化対応用のpayment-APIを叩き、card_tokenを手に入れている必要がある
    レスポンスは成功か否かのみ返す
=cut
    my ($self, $c) = @_;

    my $reservation_id = $c->req->parameters->get('reservation_id') // 0;
    my $card_token = $c->req->parameters->get('card_token') // "";

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    # 予約IDで検索
    my $query = 'SELECT * FROM reservations WHERE reservation_id=?';
    my $reservation = $dbh->select_row($query, $reservation_id);
    if (!$reservation_id) {
        warn("no reservation");
        return $self->error_with_msg($c, HTTP_NOT_FOUND, '予約情報がみつかりません');
    }

    # 支払い前のユーザチェック。本人以外のユーザの予約を支払ったりキャンセルできてはいけない。
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }
    if ($reservation->{user_id} != $user->{id}) {
        warn("not match reservation user id");
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '他のユーザIDの支払いはできません');
    }

    # 予約情報の支払いステータス確認
    if ($reservation->{status} eq "done") {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '既に支払いが完了している予約IDです');
    }

    # 決済する
    my $pay_info = {
        payment_information => {
            card_token => $card_token,
            reservation_id => number $reservation_id,
            amount => number $reservation->{amount}
        }
    };
    my $json = JSON::encode_json($pay_info);

    my $payment_api = $ENV{PAYMENT_API} // "http://payment:5000";

    my $req = HTTP::Request->new(POST => $payment_api . "/payment");
    $req->header("Content-Type", "application/json");
    $req->content($json);

    my $ua  = LWP::UserAgent->new(
        agent => "isucon9-final-webapp",
    );
    my $res = $ua->request($req);

    # リクエスト失敗
    if ($res->code != 200) {
        warn($res->status_line);
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, '決済に失敗しました。カードトークンや支払いIDが間違っている可能性があります');
    }

    my $output = eval {
        JSON::decode_json($res->content);
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'JSON parseに失敗しました');
    }

    # 予約情報の更新
    $query = 'UPDATE reservations SET status=?, payment_id=? WHERE reservation_id=?';
    $dbh->query(
        $query,
        "done",
        $output->{payment_id},
        $reservation_id
    );

    $txn->commit();

    $c->render_json({
        is_ok => JSON::true
    });
};

#mux.HandleFunc(pat.Get("/api/auth"), getAuthHandler)
get '/api/auth' => sub {
    my ($self, $c) = @_;
    #  userID取得
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    $c->render_json({
        email => $user->{email}
    });
};

#mux.HandleFunc(pat.Post("/api/auth/signup"), signUpHandler)
post '/api/auth/signup' => [qw/allow_json_request/] => sub {
=comment
    ユーザー登録
    POST /auth/signup
=cut
    my ($self, $c) = @_;

    my $email = $c->req->parameters->get('email') // "";
    my $password = $c->req->parameters->get('password') // "";

    my $pbkdf2 = Crypt::PBKDF2->new(
      hash_class => 'HMACSHA2',
      iterations => 100,
      output_len => 256,
      hash_args => { sha_size => 256 }
    );
    my $salt = Crypt::OpenSSL::Random::random_bytes(1024);
    my $super_secure_password = $pbkdf2->PBKDF2($password, $salt);

    $self->dbh->query(
        'INSERT INTO `users` (`email`, `salt`, `super_secure_password`) VALUES (?, ?, ?)',
        $email,
        $salt,
        $super_secure_password
    );

    $c->render_json({
        is_error => JSON::false,
        message => "registration complete"
    });
};

#mux.HandleFunc(pat.Post("/api/auth/login"), loginHandler)
post '/api/auth/login' => [qw/allow_json_request/] => sub {
=comment
    ログイン
    POST /auth/login
=cut
    my ($self, $c) = @_;

    my $email = $c->req->parameters->get('email') // "";
    my $password = $c->req->parameters->get('password') // "";

    my $user = $self->dbh->select_row(
        'SELECT * FROM users WHERE email=?',
        $email
    );
    if (!$user) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, 'authentication failed');
    }

    my $pbkdf2 = Crypt::PBKDF2->new(
      hash_class => 'HMACSHA2',
      iterations => 100,
      output_len => 256,
      hash_args => { sha_size => 256 }
    );

    my $hash = $pbkdf2->PBKDF2($password, $user->{salt});
    if ($hash ne $user->{super_secure_password}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, 'authentication failed');
    }

    my $session = Plack::Session->new($c->env);
    $session->set('user_id' => $user->{id});
    $session->set('csrf_token' => secure_random_str(20));

    $c->render_json({
        is_error => JSON::false,
        message => "autheticated"
    });
};

#mux.HandleFunc(pat.Post("/api/auth/logout"), logoutHandler)
post '/api/auth/logout' => sub {
=comment
    ログアウト
    POST /auth/logout
=cut
    my ($self, $c) = @_;
    my $session = Plack::Session->new($c->env);
    $session->set('user_id' => 0);
    $session->set('csrf_token' => secure_random_str(20));

    $c->render_json({
        is_error => JSON::false,
        message => "logged out"
    });
};

sub makeReservationResponse {
    my ($self,$r) = @_;
    my %res;

    my $departure = $self->dbh->select_one(
        'SELECT departure FROM train_timetable_master WHERE date=? AND train_class=? AND train_name=? AND station=?',
        $r->{date},
        $r->{train_class},
        $r->{train_name},
        $r->{departure}
    );
    if (!$departure) {
        die 'no departure';
    }

    my $arrival = $self->dbh->select_one(
        'SELECT departure FROM train_timetable_master WHERE date=? AND train_class=? AND train_name=? AND station=?',
        $r->{date},
        $r->{train_class},
        $r->{train_name},
        $r->{arrival}
    );
    if (!$arrival) {
        die 'no departure';
    }

    $res{reservation_id} = number $r->{reservation_id};
    $res{date} = $r->{date};
    $res{amount} = number $r->{amount};
    $res{adult} = number $r->{adult};
    $res{child} = number $r->{child};
    $res{departure} = $r->{departure};
    $res{arrival} = $r->{arrival};
    $res{train_class} = $r->{train_class};
    $res{train_name} = $r->{train_name};
    $res{departure_time} = $departure;
    $res{arrival_time} = $arrival;
    $res{seats} = [];

    my $query = 'SELECT * FROM seat_reservations WHERE reservation_id=?';
    my $seats = $self->dbh->select_all($query, $r->{reservation_id});
    for my $s (@$seats) {
        push @{$res{seats}}, {
            reservation_id => number $s->{reservation_id},
            car_number => number $s->{car_number},
            seat_row => number $s->{seat_row},
            seat_column => $s->{seat_column},
        };
    }

	# 1つの予約内で車両番号は全席同じ
    $res{car_number} = $res{seats}->[0]->{car_number};

	if ($res{seats}->[0]->{car_number} == 0) {
        $res{seat_class} = "non-reserved";
	}
    else {
        # 座席種別を取得
        $query = 'SELECT * FROM seat_master WHERE train_class=? AND car_number=? AND seat_column=? AND seat_row=?';
        my $seat = $self->dbh->select_row(
            $query,
            $res{train_class},
            $res{car_number},
            $res{seats}->[0]->{seat_column},
            $res{seats}->[0]->{seat_row},
        );
        if (!$seat) {
            die 'no seat';
        }
        $res{seat_class} = $seat->{seat_class};
	}

    for my $v (@{$res{seats}}) {
        delete $v->{reservation_id};
        delete $v->{car_number};
    }
	return \%res;
}


#mux.HandleFunc(pat.Get("/api/user/reservations"), userReservationsHandler)
get '/api/user/reservations' => sub {
    my ($self, $c) = @_;
    #  userID取得
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $query = 'SELECT * FROM reservations WHERE user_id=?';
    my $reservation_list = $self->dbh->select_all($query, $user->{id});

    my @reservation_response_list;
    for my $r (@$reservation_list) {
        my $res = eval {
            $self->makeReservationResponse($r);
        };
        if ($@) {
            warn("makeReservationResponse() ", $@);
            return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
        }
        push @reservation_response_list, $res;
    }

    $c->render_json(\@reservation_response_list);
};

#mux.HandleFunc(pat.Get("/api/user/reservations/:item_id"), userReservationResponseHandler)
get '/api/user/reservations/{item_id:\d+}' => sub {
    my ($self, $c) = @_;

    #  userID取得
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $item_id = $c->args->{item_id};

    my $query = 'SELECT * FROM reservations WHERE reservation_id=? AND user_id=?';
    my $reservation = $self->dbh->select_row($query, $item_id, $user->{id});
    if (!$reservation) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "Reservation not found");
    }

    my $res = eval {
        $self->makeReservationResponse($reservation);
    };
    if ($@) {
        warn("makeReservationResponse() ", $@);
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, $@);
    }

    $c->render_json($res);
};

#mux.HandleFunc(pat.Post("/api/user/reservations/:item_id/cancel"), userReservationCancelHandler)
post '/api/user/reservations/{item_id:\d+}/cancel' => sub {
    my ($self, $c) = @_;

    #  userID取得
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $item_id = $c->args->{item_id};

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $query = 'SELECT * FROM reservations WHERE reservation_id=? AND user_id=?';
    my $reservation = $dbh->select_row($query, $item_id, $user->{id});
    warnf("CANCEL %s %s %s", $reservation, $item_id, $user->{id});
    if (!$reservation) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "reservations naiyo");
    }

    if ($reservation->{status} eq "rejected") {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "何らかの理由により予約はRejected状態です");
    }
    elsif ($reservation->{status} eq "done") {
        # 支払いをキャンセルする
        my $pay_info = {
            payment_id => $reservation->{payment_id}
        };
        my $json = JSON::encode_json($pay_info);

        my $payment_api = $ENV{PAYMENT_API} // "http://payment:5000";

        my $req = HTTP::Request->new(DELETE => $payment_api . "/payment/".$reservation->{payment_id});
        $req->header("Content-Type", "application/json");
        $req->content($json);

        my $ua  = LWP::UserAgent->new(
            agent => "isucon9-final-webapp",
        );
        my $res = $ua->request($req);

        # リクエスト失敗
        if ($res->code != 200) {
            warn($res->status_line);
            return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "決済に失敗しました。支払いIDが間違っている可能性があります");
        }

        my $output = eval {
            JSON::decode_json($res->content);
        };
        if ($@) {
            warn $@;
            return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'JSON parseに失敗しました');
        }

        warnf($output);

    }
    else {
        # pass(requesting状態のものはpayment_id無いので叩かない)
    }

    $query = 'DELETE FROM reservations WHERE reservation_id=? AND user_id=?';
    $dbh->query($query, $item_id, $user->{id});

    $query = 'DELETE FROM seat_reservations WHERE reservation_id=?';
    my $rows = $dbh->query($query, $item_id);
    if ($rows == 0) {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'seat naiyo');
    }

    $txn->commit();
    $c->render_json({
        is_error => JSON::false,
        message => "cancell complete"
    });
};

1;
