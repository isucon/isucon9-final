<?php


namespace App;

use Composer\XdebugHandler\Status;
use DateTime;
use GuzzleHttp\Client;
use GuzzleHttp\Exception\RequestException;
use PDO;
use Psr\Container\ContainerInterface;
use Psr\Http\Message\UploadedFileInterface;
use Psr\Log\LoggerInterface;
use Slim\Http\Request;
use Slim\Http\Response;
use Slim\Http\StatusCode;

class Service
{
    /**
     * @var LoggerInterface
     */
    private $logger;

    /**
     * @var PDO
     */
    private $dbh;

    /**
     * @var array
     */
    private $settings;

    const AVAILABLE_DAYS = 10;

    const TRAIN_CLASS_MAP = [
        'express' => '最速',
        'semi_express' => '中間',
        'local' => '遅いやつ',
        ];

    // constructor receives container instance
    public function __construct(ContainerInterface $container)
    {
        $this->logger = $container->get('logger');
        $this->dbh = $container->get('dbh');
        $this->settings = $container->get('settings');
    }


    // utils
    private function errorResponse($message)
    {
        return [
            'is_error' => true,
            'message' => $message,
        ];
    }

    private function checkAvailableDate(DateTime $date): bool
    {
        $base = new DateTime('2020-01-01 00:00:00');
        $interval = new DateInterval(sprintf('P%dD', self::AVAILABLE_DAYS));
        $base->add($interval);
        return $base > $date;
    }

    private function getUsableTrainClassList(array $fromStation, array $toStation): array
    {
        $usable = [];
        foreach (self::TRAIN_CLASS_MAP as $k => $v) {
            $usable[$k] = $v;
        }

        // TODO check valid ops
        if (! $fromStation['is_stop_express']) {
            unset($usable['express']);
        }

        if (! $fromStation['is_stop_semi_express']) {
            unset($usable['semi_express']);
        }

        if (! $fromStation['is_stop_local']) {
            unset($usable['local']);
        }


        // TODO check valid ops
        if (! $toStation['is_stop_express']) {
            unset($usable['express']);
        }

        if (! $toStation['is_stop_semi_express']) {
            unset($usable['semi_express']);
        }

        if (! $toStation['is_stop_local']) {
            unset($usable['local']);
        }

        return array_values($usable);
    }

    private function getAvailableSeats(array $train, array $fromStation, array $toStation, string $seatClass, bool $isSmokingSeat): array
    {
        // 全ての座席を取得する
        $stmt = $this->dbh->prepare("SELECT * FROM seat_master WHERE train_class=? AND seat_class=? AND is_smoking_seat=?;");
        $stmt->execute([
            $train['train_class'],
            $seatClass,
            $isSmokingSeat,
        ]);
        $seatList = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($seatList === false) {
            // TODO Error
            return [];
        }
        $availableSeatMap = [];
        foreach ($seatList as $k => $seat) {
            $key = sprintf("%d_%d_%s", $seat['car_number'], $seat['seat_row'], $seat['seat_column']);
            $availableSeatMap[$key] = $seat;
        }

        // すでに取られている予約を取得する
        $query = "SELECT `sr`.`reservation_id`, `sr`.`car_number`, `sr`.`seat_row`, `sr`.`seat_column` " .
            "FROM `seat_reservations` sr, `reservations` r, `seat_master` s, `station_master` std, `station_master` sta " .
            "WHERE " .
            "r.reservation_id=sr.reservation_id AND " .
            "s.train_class=r.train_class AND " .
            "s.car_number=sr.car_number AND " .
            "s.seat_column=sr.seat_column AND " .
            "s.seat_row=sr.seat_row AND " .
            "std.name=r.departure AND" .
            "sta.name=r.arrival ";
        if ($train['is_nobori']) {
            $query .= "AND ((sta.id < ? AND ? <= std.id) OR (sta.id < ? AND ? <= std.id) OR (? < sta.id AND std.id < ?))";
        } else {
            $query .= "AND ((std.id <= ? AND ? < sta.id) OR (std.id <= ? AND ? < sta.id) OR (sta.id < ? AND ? < std.id))";
        }
        $stmt = $this->dbh->prepare($query);
        $stmt->execute([
            $fromStation['id'],
            $fromStation['id'],
            $toStation['id'],
            $toStation['id'],
            $fromStation['id'],
            $toStation['id'],
        ]);
        $seatReservationList = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($seatReservationList === false) {
            // TODO Error
            return [];
        }

        foreach ($seatReservationList as $seatReservation) {
            $key = sprintf("%d_%d_%s", $seatReservation['car_number'], $seatReservation['seat_row'], $seatReservation['seat_column']);
            unset($availableSeatMap[$key]);
        }

        return array_values($availableSeatMap);
    }

    private function fareCalc(DateTime $date, int $depStation, int $destStation, string $trainClass, string $seatClass): int
    {
        // 料金計算メモ
        // 距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)
        $sql = "SELECT * FROM station_master WHERE id=?";
        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([$depStation]);
        $fromStation = $stmt->fetch(PDO::FETCH_ASSOC);
        // TODO Error

        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([$destStation]);
        $toStation = $stmt->fetch(PDO::FETCH_ASSOC);
        // TODO Error

        $distFare = $this->getDistanceFare(abs($toStation['distance'] - $fromStation['distance']));

        // 期間・車両・座席クラス倍率
        $stmt = $this->dbh->prepare("SELECT * FROM `fare_master` WHERE `train_class`=? AND `seat_class`=? ORDER BY `start_date`");
        $stmt->execute([
            $trainClass,
            $seatClass,
        ]);
        $fareList = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($fareList === false) {
            // fare_master does not exists
        }

        $selectedFare = $fareList[0];
        foreach($fareList as $fare) {
            if ($fare['start_date'] < $date) {
                $selectedFare = $fare;
            }
        }
        return (int) ($distFare * $selectedFare['fare_multiplier']);
    }

    protected function getDistanceFare(float $origToDestDistance): int
    {
        $stmt = $this->dbh->prepare("SELECT `distance`,`fare` FROM `distance_fare_master` ORDER BY `distance`");
        $stmt->execute([]);
        $distanceFareList = $stmt->fetchAll(PDO::FETCH_ASSOC);

        $lastDistance = 0.0;
        $lastFare = 0;
        foreach ($distanceFareList as $distanceFare) {
            if (($lastDistance < $origToDestDistance) && ($origToDestDistance < $distanceFare[''])) {
                break;
            }
            $lastDistance = $distanceFare['distance'];
            $lastFare = $distanceFare['fare'];
        }
        return $lastFare;
    }

    private function jsonPayload(Request $request): array  {
        $data = json_decode($request->getBody());
        if (JSON_ERROR_NONE !== json_last_error()) {
            throw new \InvalidArgumentException(json_last_error_msg());
        }
        return $data;
    }


    // handler

    public function getStationsHandler(Request $request, Response $response, array $args)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `station_master` BY `id`');
        $sth->execute();
        $data = $sth->fetchAll(PDO::FETCH_ASSOC);
        if ($data === false) {
            return $response->withJson($this->errorResponse($sth->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
        }

        $station = [];
        foreach ($data as $elem) {
            unset($elem['distance']);
            $station[] = $elem;
        }
        return $response->withJson($station);
    }

    public function trainSearchHandler(Request $request, Response $response, array  $args)
    {
        $date = $request->getParam('use_at');
        try {
            $dt = new DateTime($date);
        } catch (\Exception $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }
        if (! $this->checkAvailableDate($dt)) {
            return $response->withJson($this->errorResponse("予約可能期間外です"), StatusCode::HTTP_NOT_FOUND);
        }

        $trainClass = $request->getParam('train_class', '');
        $fromName = $request->getParam('from', '');
        $toName = $request->getParam('to', '');
        $adult = $request->getParam('adult', '');
        $child = $request->getParam('child', '');

        try {
            $sql = "SELECT * FROM station_master WHERE name=?";
            $sth = $this->dbh->prepare($sql);
            $sth->execute([$fromName]);
            $fromStation = $sth->fetch(PDO::FETCH_ASSOC);
            if ($fromStation === false) {
                return $response->withJson($this->errorResponse(['not found']), StatusCode::HTTP_BAD_REQUEST);
            }

            $sth = $this->dbh->prepare($sql);
            $sth->execute([$toName]);
            $toStation = $sth->fetch(PDO::FETCH_ASSOC);
            if ($toStation === false) {
                return $response->withJson($this->errorResponse(['not found']), StatusCode::HTTP_BAD_REQUEST);
            }
            $isNobori = false;
            if ($fromStation['distance'] > $toStation['distance']) {
                $isNobori = true;
            }

            $usableTrainClassList = $this->getUsableTrainClassList($fromStation, $toStation);

            if ($trainClass === '') {
                $in = str_repeat('?,', count($usableTrainClassList) -1) .  '?';
                $sql = "SELECT * FROM `train_master` WHERE date=? AND `train_class` IN (${in}) AND `is_nobori`=?";
                $args = array_merge([
                    [$date],
                    $usableTrainClassList,
                    [$isNobori],
                ]);
            } else {
                $in = str_repeat('?,', count($usableTrainClassList) -1) .  '?';
                $sql = "SELECT * FROM `train_master` WHERE date=? AND `train_class` IN (${in}) AND `is_nobori`=? AND `train_class`=?";
                $args = array_merge([
                    [$date],
                    $usableTrainClassList,
                    [$isNobori],
                    [$trainClass],
                ]);
            }
            $sth = $this->dbh->prepare($sql);
            $sth->execute($args);
            $trainList = $sth->fetchAll(PDO::FETCH_ASSOC);
            if ($trainClass === false) {
                return $response->withJson($this->errorResponse(['not found']), StatusCode::HTTP_BAD_REQUEST);
            }

            $sql = "SELECT * FROM station_master ORDER BY distance";
            if ($isNobori) {
                // if nobori reverse the order
                $sql = $sql . " DESC";
            }

            $stations = $this->dbh->exec($sql);
            if ($stations === false) {
                return $response->withJson($this->errorResponse(['not found']), StatusCode::HTTP_BAD_REQUEST);
            }

            $this->logger->info("From:", [$fromStation]);
            $this->logger->info("To:", [$toStation]);

            $trainSearchResponseList = [];

            foreach ($trainList as $k => $train) {
                $isSeekedToFirstStation = false;
                $isContainsOriginStation = false;
                $isContainsDestStation = false;
                $i = 0;
                foreach ($stations as $s => $station) {
                    // 駅リストを列車の発駅まで読み飛ばして頭出しをする
                    // 列車の発駅以前は止まらないので無視して良い
                    if (! $isSeekedToFirstStation) {
                        if ($station['name'] === $train['start_station']) {
                            $isSeekedToFirstStation = true;
                        } else {
                            continue;
                        }
                    }

                    if ($station['id'] === $fromStation['id']) {
                        // 発駅を経路中に持つ編成の場合フラグを立てる
                        $isContainsOriginStation = true;
                    }

                    if ($station['id'] === $toStation['id']) {
                        if ($isContainsOriginStation) {
                            $isContainsDestStation = true;
                            break;
                        } else {
                            // 出発駅より先に終点が見つかったとき
                            $this->logger->info("なんかおかしい");
                            break;
                        }
                    }

                    if ($station['name'] === $train['last_station']) {
                        // 駅が見つからないまま当該編成の終点に着いてしまったとき
                        break;
                    }
                    $i++;
                }

                if ($isContainsOriginStation && $isContainsDestStation) {
                    // 列車情報

                    // 所要時間
                    $sql = "SELECT `departure` FROM `train_timetable_master` WHERE `date`=? AND `train_class`=? AND `train_name`=? AND `station`=?";
                    $stmt = $this->dbh->prepare($sql);
                    $departure = $stmt->execute([
                        $date,
                        $trainClass,
                        $train['name'],
                        $fromStation['name']
                    ]);
                    if ($departure === false) {
                        return $response->withJson($this->errorResponse(['failed to query']), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                    }
                    // TODO check unneed
                    $departureDate = DateTime($departure);
                    if ($date > $departureDate) {
                        // 乗りたい時刻より出発時刻が前なので除外
                        continue;
                    }

                    $sth = $this->dbh->prepare("SELECT `arrival` FROM `train_timetable_master` WHERE `date`=? AND `train_class`=? AND `train_name`=? AND `station`=?");
                    $arrival = $sth->execute([
                        $date,
                        $trainClass,
                        $train['name'],
                        $toStation['name']
                    ]);
                    if ($arrival === false) {
                        return $response->withJson($this->errorResponse(['failed to query']), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                    }

                    $premium_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'premium', false);
                    $premium_smoke_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'premium', true);
                    $reserved_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'reserved', false);
                    $reserved_smoke_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'reserved', true);

                    $premium_avail = $premium_smoke_avail = $reserved_avail = $reserved_smoke_avail = "○";
                    if (count($premium_avail_seats) == 0) {
                        $premium_avail = "×";
                    } elseif (count($premium_avail_seats) < 10) {
                        $premium_avail = "△";
                    }

                    if (count($premium_smoke_avail_seats) == 0) {
                        $premium_smoke_avail = "×";
                    } elseif (count($premium_smoke_avail_seats) < 10) {
                        $premium_smoke_avail = "△";
                    }

                    if (count($reserved_avail_seats) == 0) {
                        $reserved_avail = "×";
                    } elseif (count($reserved_avail_seats) < 10) {
                        $reserved_avail = "△";
                    }

                    if (count($reserved_smoke_avail) == 0) {
                        $reserved_smoke_avail = "×";
                    } elseif (count($reserved_smoke_avail) < 10) {
                        $reserved_smoke_avail = "△";
                    }

                    $seatAvailability = [
                        "premium" =>        $premium_avail,
                        "premium_smoke" =>   $premium_smoke_avail,
                        "reserved" =>       $reserved_avail,
                        "reserved_smoke" =>  $reserved_smoke_avail,
                        "non_reserved" =>   "○",
                    ];

                    // 料金計算
                    $premiumFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $trainClass, "premium");
                    $premiumFare = ($premiumFare*$adult) + (($premiumFare/2)*$child) ;
                    $reservedFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $trainClass, "reserved");
                    $reservedFare = ($reservedFare * $adult) + (($reservedFare/2)*$child);
                    $nonReservedFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $trainClass, "non-reserved");
                    $nonReservedFare = ($nonReservedFare * $adult) + (($nonReservedFare/2) *$child);

                    $fareInformation = [
                        "premium" => int($premiumFare),
                        "premium_smoke" => int($premiumFare),
                        "reserved" => int($reservedFare),
                        "reserved_smoke" => int($reservedFare),
                        "non_reserved" => int($nonReservedFare),
                    ];

                    $trainSearchResponseList[] = [
                        "train_class" => $train['train_class'],
                        "train_name" => $train['train_name'],
                        "start" => $train['start_station'],
                        "last" => $train['last_station'],
                        "departure" => $fromStation['name'],
                        "arrival" => $toStation['name'],
                        "departure_time"=> $departure,
                        "arrival_time" => $arrival,
                        "seat_availability" => $seatAvailability,
                        "seat_fare" => $fareInformation,
                    ];

                    if (count($trainSearchResponseList) >= 10) {
                        break;
                    }
                }
            }
        } catch (\PDOException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }
        return $response->withJson($trainSearchResponseList);
    }

    public function trainSeatsHandler(Request $request, Response $response, array $args)
    {
        $date = new DateTime($request->getParam("date", ""));
        if (! $this->checkAvailableDate($date)) {
            return $response->withJson($this->errorResponse("予約可能期間外です"), StatusCode::HTTP_NOT_FOUND);
        }

        $trainClass = $request->getParam('train_class', '');
        $trainName = $request->getParam('train_name', '');
        $carNumber = $request->getParam('car_number', '');
        $fromName = $request->getParam('from', '');
        $toName = $request->getParam('to', '');

        // 対象列車の取得
        $stmt = $this->dbh->prepare("SELECT * FROM `train_master` WHERE `date`=? AND `train_class`=? AND `train_name`=?");
        $stmt->execute([
            $date,
            $trainClass,
            $trainName
        ]);
        $train = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($train === false) {
            return $response->withJson($this->errorResponse("列車が存在しません"), StatusCode::HTTP_NOT_FOUND);
        }

        $sql = "SELECT * FROM `station_master` WHERE `name`=?";
        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([
            $fromName
        ]);
        $fromStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($fromStation === false) {
            return $response->withJson($this->errorResponse("fromStation: no rows"), StatusCode::HTTP_BAD_REQUEST);
        }

        $sql = "SELECT * FROM `station_master` WHERE `name`=?";
        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([
            $toName,
        ]);
        $toStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($toStation === false) {
            return $response->withJson($this->errorResponse("toStation: no rows"), StatusCode::HTTP_BAD_REQUEST);
        }

        $usableTrainClassList = $this->getUsableTrainClassList($fromStation, $toStation);
        $usable = in_array($train['train_class'], $usableTrainClassList);

        if (! $usable) {
            return $response->withJson($this->errorResponse("invalid train_class"), StatusCode::HTTP_BAD_REQUEST);
        }

        $stmt = $this->dbh->prepare("SELECT * FROM `seat_master` WHERE `train_class`=? AND `car_number`=? ORDER BY `seat_row`, `seat_column`");
        $stmt->execute([
            $trainClass,
            $carNumber,
        ]);
        $seatList = $stmt->fetchAll(PDO::FETCH_ASSOC);

        $seatInformationList = [];
        foreach ($seatList as $seat) {
            $s = [
                'row' => $seat['seat_row'],
                'column' => $seat['seat_column'],
                'class' => $seat['seat_class'],
                'is_smoking_seat' => $seat['is_smoking_seat'],
                'is_occupied' => false,
            ];
            $stmt = $this->dbh->prepare("SELECT s.* FROM `seat_reservations` s, `reservations` r WHERE `r`.`date`=? AND `r`.`train_class`=? AND `r`.`train_name`=? AND `car_number`=? AND `seat_row`=? AND `seat_column`=?");
            $stmt->execute([
                $date,
                $seat['train_class'],
                $trainName,
                $seat['car_number'],
                $seat['seat_row'],
                $seat['seat_column'],
            ]);
            $seatReservationList = $stmt->fetchAll(PDO::FETCH_ASSOC);
            if ($seatReservationList === false) {
                return $response->withJson($this->errorResponse("failed to fetch seat_reservations"), StatusCode::HTTP_BAD_REQUEST);
            }
            foreach ($seatReservationList as $seatReservation) {
                $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `reservation_id`=?");
                $stmt->execute([$seatReservation['reservation_id']]);
                $reservation = $stmt->fetch(PDO::FETCH_ASSOC);
                if ($reservation) {
                    return $response->withJson($this->errorResponse("failed to fetch seat_reservations"), StatusCode::HTTP_BAD_REQUEST);
                }

                $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
                $stmt->execute([$reservation['departure']]);
                $departureStation = $stmt->fetch(PDO::FETCH_ASSOC);
                if ($departureStation === false) {
                    return $response->withJson($this->errorResponse("failed to fetch departure"), StatusCode::HTTP_BAD_REQUEST);
                }

                $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
                $stmt->execute([$reservation['arrival']]);
                $arrivalStation = $stmt->fetch(PDO::FETCH_ASSOC);
                if ($departureStation === false) {
                    return $response->withJson($this->errorResponse("failed to fetch arrivalStation"), StatusCode::HTTP_BAD_REQUEST);
                }

                if ($train['is_nobori']) {
                    if (($toStation['id'] < $arrivalStation['id']) && $fromStation['id'] <= $arrivalStation['id']) {
                        // pass
                    } elseif (($toStation['id'] >= $departureStation['id']) && $fromStation['id'] > $departureStation['id']) {
                        // pass
                    } else {
                        $s['is_occupied'] = true;
                    }
                } else {
                    if (($fromStation['id'] < $departureStation['id']) && $toStation['id'] <= $departureStation['id']) {
                        // pass
                    } elseif (($fromStation['id'] >= $arrivalStation['id']) && $toStation['id'] > $arrivalStation['id']) {
                        // pass
                    } else {
                        $s['is_occupied'] = true;
                    }
                }
            }
            $seatInformationList[] = $s;
        }

        // 各号車の情報
        $simpleCarInformationList = [];
        $i = 1;
        while (true) {
            $stmt = $this->dbh->query("SELECT * FROM `seat_master` WHERE `train_class`=? AND `car_number`=? ORDER BY `seat_row`, `seat_column` LIMIT 1");
            $stmt->execute([
                $trainClass,
                $i,
            ]);
            $seat = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($seat === false) {
                break;
            }
            $simpleCarInformationList[] = [
                'car_number' => $i,
                'seat_class' => $seat['seat_class'],
            ];
            $i++;
        }

        $carInformation = [
            'date' => $date->format('Y/m/d'),
            'train_class' => $trainClass,
            'train_name' => $trainName,
            'car_number' => $carNumber,
            'seats' => $seatInformationList,
            'cars' => $simpleCarInformationList,
        ];

        return $response->withJson($carInformation);
    }

    public function trainReservationHandler(Request $request, Response $response, array $args)
    {
        // request payload
        /**
         * 	          string        `json:"date"`
        *     string        `json:"train_name"`
        *    string        `json:"train_class"`
        *     int           `json:"car_number"`
        * bool          `json:"is_smoking_seat"`
        *     string        `json:"seat_class"`
        *     string        `json:"departure"`
        *       string        `json:"arrival"`
        *         int           `json:"child"`
        *         int           `json:"adult"`
        *        string        `json:"Column"`
        *         []{int row, string column} `json:"seats"`
         */
    }

    public function initialize(Request $request, Response $response, array $args)
    {
        return $response->withJson(["language" => "php"]);
    }

    public function settingsHandler(Request $request, Response $response, array $args)
    {
        return $response->withJson(["payment_api" => Environment::get('PAYMENT_API', 'http://localhost:5000')]);
    }
}
