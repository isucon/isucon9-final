<?php


namespace App;

use DateTime;
use DateTimeZone;
use GuzzleHttp\Client;
use GuzzleHttp\Exception\RequestException;
use PDO;
use Psr\Container\ContainerInterface;
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
     * @var \SlimSession\Helper
     */
    private $session;

    private const AVAILABLE_DAYS = 10;

    private const TRAIN_CLASS_MAP = [
        'express' => '最速',
        'semi_express' => '中間',
        'local' => '遅いやつ',
        ];

    private const DATE_SQL_FORMAT = 'Y-m-d';

    // constructor receives container instance
    public function __construct(ContainerInterface $container)
    {
        $this->logger = $container->get('logger');
        $this->dbh = $container->get('dbh');
        $this->session = $container->get('session');
    }


    // utils
    private function messageResponse($message)
    {
        return [
            'is_error' => false,
            'message' => $message,
        ];
    }

    private function errorResponse($message)
    {
        if (is_array($message)) {
            $message = join(" ", $message);
        }
        return [
            'is_error' => true,
            'message' => $message,
        ];
    }

    private function checkAvailableDate(DateTime $date): bool
    {
        $base = new DateTime('2020-01-01 00:00:00');
        $interval = new \DateInterval(sprintf('P%dD', self::AVAILABLE_DAYS));
        $base->add($interval);
        return $base > $date;
    }

    private function getUsableTrainClassList(array $fromStation, array $toStation): array
    {
        $usable = [];
        foreach (self::TRAIN_CLASS_MAP as $k => $v) {
            $usable[$k] = $v;
        }

        if (! (bool) $fromStation['is_stop_express']) {
            unset($usable['express']);
        }

        if (! (bool) $fromStation['is_stop_semi_express']) {
            unset($usable['semi_express']);
        }

        if (! (bool) $fromStation['is_stop_local']) {
            unset($usable['local']);
        }


        if (! (bool) $toStation['is_stop_express']) {
            unset($usable['express']);
        }

        if (! (bool) $toStation['is_stop_semi_express']) {
            unset($usable['semi_express']);
        }

        if (! (bool) $toStation['is_stop_local']) {
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
            throw new \PDOException("not found");
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
            "std.name=r.departure AND " .
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
            throw new \PDOException("not found");
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
        $sql = "SELECT * FROM `station_master` WHERE id=?";
        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([$depStation]);
        $fromStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($fromStation === false) {
            throw new \PDOException('not found');
        }

        $stmt = $this->dbh->prepare($sql);
        $stmt->execute([$destStation]);
        $toStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($toStation === false) {
            throw new \PDOException('not found');
        }

        $distFare = $this->getDistanceFare(abs($toStation['distance'] - $fromStation['distance']));

        // 期間・車両・座席クラス倍率
        $stmt = $this->dbh->prepare("SELECT * FROM `fare_master` WHERE `train_class`=? AND `seat_class`=? ORDER BY `start_date`");
        $stmt->execute([
            $trainClass,
            $seatClass,
        ]);
        $fareList = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($fareList === false) {
            throw new \PDOException('not found');
        }

        $selectedFare = $fareList[0];
        foreach ($fareList as $fare) {
            $dt = new \DateTime($fare['start_date']);
            if ($dt < $date) {
                $selectedFare = $fare;
            }
        }
        return (int) ($distFare * $selectedFare['fare_multiplier']);
    }

    private function getDistanceFare(float $origToDestDistance): int
    {
        $stmt = $this->dbh->prepare("SELECT `distance`,`fare` FROM `distance_fare_master` ORDER BY `distance`");
        $stmt->execute([]);
        $distanceFareList = $stmt->fetchAll(PDO::FETCH_ASSOC);

        $lastDistance = 0.0;
        $lastFare = 0;
        foreach ($distanceFareList as $distanceFare) {
            if (($lastDistance < $origToDestDistance) && ($origToDestDistance < $distanceFare['distance'])) {
                break;
            }
            $lastDistance = $distanceFare['distance'];
            $lastFare = $distanceFare['fare'];
        }
        return $lastFare;
    }

    private function jsonPayload(Request $request): array
    {
        $data = json_decode($request->getBody(), true);
        if (JSON_ERROR_NONE !== json_last_error()) {
            throw new \InvalidArgumentException(json_last_error_msg());
        }
        return $data;
    }

    private function getUser():array
    {
        if (! $this->session->exists('user_id')) {
            throw new \DomainException('no session');
        }

        $user_id = $this->session->get('user_id');
        $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
        $sth->execute([$user_id]);
        $user = $sth->fetch(PDO::FETCH_ASSOC);
        if ($user === false) {
            throw new \DomainException('user not found');
        }
        return $user;
    }

    private function makeReservationResponse(array $reservation): array
    {
        /**
         * int               `json:"reservation_id"`
         * string            `json:"date"`
         * string            `json:"train_class"`
         * string            `json:"train_name"`
         * int               `json:"car_number"`
         * string            `json:"seat_class"`
         * int               `json:"amount"`
         * int               `json:"adult"`
         * int               `json:"child"`
         * string            `json:"departure"`
         * string            `json:"arrival"`
         * string            `json:"departure_time"`
         * string            `json:"arrival_time"`
         * []SeatReservation `json:"seats"`
         */
        $rtn = [
            'reservation_id' => $reservation['reservation_id'],
            'date' => (new \DateTime($reservation['date']))->format('Y/m/d'),
            'amount' => $reservation['amount'],
            'adult' => $reservation['adult'],
            'child' => $reservation['child'],
            'departure' => $reservation['departure'],
            'arrival' => $reservation['arrival'],
            'train_class' => $reservation['train_class'],
            'train_name' => $reservation['train_name'],
        ];
        $stmt = $this->dbh->prepare("SELECT `departure` FROM `train_timetable_master` WHERE `date`=? AND `train_class`=? AND `train_name`=? AND `station`=?");
        $stmt->execute([
            (new \DateTime($reservation['date']))->format(self::DATE_SQL_FORMAT),
            $reservation['train_class'],
            $reservation['train_name'],
            $reservation['departure'],
        ]);
        $departure = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($departure === false) {
            throw new \DomainException();
        }
        $rtn['departure_time'] = $departure['departure'];

        $stmt = $this->dbh->prepare("SELECT `arrival` FROM `train_timetable_master` WHERE `date`=? AND `train_class`=? AND `train_name`=? AND `station`=?");
        $stmt->execute([
            (new \DateTime($reservation['date']))->format(self::DATE_SQL_FORMAT),
            $reservation['train_class'],
            $reservation['train_name'],
            $reservation['arrival'],
        ]);
        $arrival = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($arrival === false) {
            throw new \DomainException();
        }
        $rtn['arrival_time'] = $arrival['arrival'];

        $stmt = $this->dbh->prepare("SELECT * FROM `seat_reservations` WHERE `reservation_id`=?");
        $stmt->execute([$reservation['reservation_id']]);
        $rtn['seats'] = $stmt->fetchAll(PDO::FETCH_ASSOC);

        // 1つの予約内で車両番号は全席同じ
        $rtn['car_number'] = $rtn['seats'][0]['car_number'];

        if ($rtn['seats'][0]['car_number'] === 0) {
            $rtn['seat_class'] = 'non-reserved';
        } else {
            $stmt = $this->dbh->prepare("SELECT * FROM `seat_master` WHERE `train_class`=? AND `car_number`=? AND `seat_column`=? AND `seat_row`=?");
            $stmt->execute([
                $reservation['train_class'],
                $rtn['car_number'],
                $rtn['seats'][0]['seat_column'],
                $rtn['seats'][0]['seat_row'],
            ]);
            $seat = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($seat === false) {
                throw new \DomainException();
            }
            $rtn['seat_class'] = $seat['seat_class'];
        }

        foreach ($rtn['seats'] as $key => $v) {
            // omit
            $rtn['seats'][$key] = [
                'reservation_id'=> 0,
                'car_number' => 0,
            ];
        }

        return $rtn;
    }



    // handler

    public function getStationsHandler(Request $request, Response $response, array $args)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `station_master` ORDER BY `id`');
        $sth->execute();
        $data = $sth->fetchAll(PDO::FETCH_ASSOC);
        if ($data === false) {
            return $response->withJson($this->errorResponse($sth->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
        }

        $station = [];
        foreach ($data as $elem) {
            unset($elem['distance']);
            $elem['is_stop_express'] = (bool) $elem['is_stop_express'];
            $elem['is_stop_semi_express'] = (bool) $elem['is_stop_semi_express'];
            $elem['is_stop_local'] = (bool) $elem['is_stop_local'];
            $station[] = $elem;
        }
        return $response->withJson($station);
    }

    public function trainSearchHandler(Request $request, Response $response, array  $args)
    {
        try {
            $date = DateTime::createFromFormat('Y-m-d\TH:i:s.vO', $request->getParam("use_at", ""));
            if (! $date) {
                $date = DateTime::createFromFormat('Y-m-d\TH:i:sO', $request->getParam("use_at", ""));
            }
            $date = $date->setTimezone(new DateTimeZone("Asia/Tokyo"));
        } catch (\Exception $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }
        if (! $this->checkAvailableDate($date)) {
            return $response->withJson($this->errorResponse("予約可能期間外です"), StatusCode::HTTP_NOT_FOUND);
        }

        $trainClass = $request->getParam('train_class', '');
        $fromName = $request->getParam('from', '');
        $toName = $request->getParam('to', '');
        $adult = $request->getParam('adult', 0);
        $child = $request->getParam('child', 0);

        try {
            $sql = "SELECT * FROM `station_master` WHERE `name`=?";
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
            $in = str_repeat('?,', count($usableTrainClassList) -1) .  '?';
            if ($trainClass === '') {
                $sql = "SELECT * FROM `train_master` WHERE `date`=? AND `train_class` IN (${in}) AND `is_nobori`=?";
                $args = array_merge(
                    [$date->format(self::DATE_SQL_FORMAT)],
                    $usableTrainClassList,
                    [$isNobori]
                );
            } else {
                $sql = "SELECT * FROM `train_master` WHERE `date`=? AND `train_class` IN (${in}) AND `is_nobori`=? AND `train_class`=?";
                $args = array_merge(
                    [$date->format(self::DATE_SQL_FORMAT)],
                    $usableTrainClassList,
                    [$isNobori],
                    [$trainClass]
                );
            }
            $sth = $this->dbh->prepare($sql);
            $sth->execute($args);
            $trainList = $sth->fetchAll(PDO::FETCH_ASSOC);
            if ($trainList === false) {
                return $response->withJson($this->errorResponse(['not found']), StatusCode::HTTP_BAD_REQUEST);
            }

            $sql = "SELECT * FROM `station_master` ORDER BY `distance`";
            if ($isNobori) {
                // if nobori reverse the order
                $sql = $sql . " DESC";
            }

            $stmt = $this->dbh->prepare($sql);
            $stmt->execute([]);
            $stations = $stmt->fetchAll(PDO::FETCH_ASSOC);

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
                    $stmt->execute([
                        $date->format(self::DATE_SQL_FORMAT),
                        $train['train_class'],
                        $train['train_name'],
                        $fromStation['name']
                    ]);
                    $departure = $stmt->fetch(PDO::FETCH_ASSOC);
                    if ($departure === false) {
                        $this->logger->error($this->dbh->errorCode(), $this->dbh->errorInfo());
                        return $response->withJson($this->errorResponse(['failed to query']), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                    }
                    $departureDate = new \DateTime($date->format('Y-m-d ') . $departure['departure']);
                    if ($date > $departureDate) {
                        // 乗りたい時刻より出発時刻が前なので除外
                        continue;
                    }

                    $stmt = $this->dbh->prepare("SELECT `arrival` FROM `train_timetable_master` WHERE `date`=? AND `train_class`=? AND `train_name`=? AND `station`=?");
                    $stmt->execute([
                        $date->format(self::DATE_SQL_FORMAT),
                        $train['train_class'],
                        $train['train_name'],
                        $toStation['name']
                    ]);
                    $arrival = $stmt->fetch(PDO::FETCH_ASSOC);
                    if ($arrival === false) {
                        return $response->withJson($this->errorResponse(['failed to query']), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                    }

                    try {
                        $premium_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'premium', false);
                        $premium_smoke_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'premium', true);
                        $reserved_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'reserved', false);
                        $reserved_smoke_avail_seats = $this->getAvailableSeats($train, $fromStation, $toStation, 'reserved', true);
                    } catch (\PDOException $e) {
                        return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
                    }

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

                    if (count($reserved_smoke_avail_seats) == 0) {
                        $reserved_smoke_avail = "×";
                    } elseif (count($reserved_smoke_avail_seats) < 10) {
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
                    $premiumFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $train['train_class'], "premium");
                    $premiumFare = ($premiumFare*$adult) + (($premiumFare/2)*$child) ;
                    $reservedFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $train['train_class'], "reserved");
                    $reservedFare = ($reservedFare * $adult) + (($reservedFare/2)*$child);
                    $nonReservedFare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $train['train_class'], "non-reserved");
                    $nonReservedFare = ($nonReservedFare * $adult) + (($nonReservedFare/2) *$child);

                    $fareInformation = [
                        "premium" => (int) $premiumFare,
                        "premium_smoke" => (int) $premiumFare,
                        "reserved" => (int) $reservedFare,
                        "reserved_smoke" => (int) $reservedFare,
                        "non_reserved" => (int) $nonReservedFare,
                    ];

                    $trainSearchResponseList[] = [
                        "train_class" => $train['train_class'],
                        "train_name" => $train['train_name'],
                        "start" => $train['start_station'],
                        "last" => $train['last_station'],
                        "departure" => $fromStation['name'],
                        "arrival" => $toStation['name'],
                        "departure_time"=> $departure['departure'],
                        "arrival_time" => $arrival['arrival'],
                        "seat_availability" => $seatAvailability,
                        "seat_fare" => $fareInformation,
                    ];

                    if (count($trainSearchResponseList) >= 10) {
                        break;
                    }
                }
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage(), [$e]);
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }
        return $response->withJson($trainSearchResponseList);
    }

    public function trainSeatsHandler(Request $request, Response $response, array $args)
    {
        try {
            $date = DateTime::createFromFormat('Y-m-d\TH:i:s.vO', $request->getParam("date", ""));
            if (! $date) {
                $date = DateTime::createFromFormat('Y-m-d\TH:i:sO', $request->getParam("date", ""));
            }
            $date = $date->setTimezone(new DateTimeZone("Asia/Tokyo"));
        } catch (\Exception $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }
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
            $date->format(self::DATE_SQL_FORMAT),
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
                'is_smoking_seat' => (bool) $seat['is_smoking_seat'],
                'is_occupied' => false,
            ];
            $stmt = $this->dbh->prepare("SELECT `s`.* FROM `seat_reservations` s, `reservations` r WHERE `r`.`date`=? AND `r`.`train_class`=? AND `r`.`train_name`=? AND `car_number`=? AND `seat_row`=? AND `seat_column`=?");
            $stmt->execute([
                $date->format(self::DATE_SQL_FORMAT),
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
                if ($reservation === false) {
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
            $stmt = $this->dbh->prepare("SELECT * FROM `seat_master` WHERE `train_class`=? AND `car_number`=? ORDER BY `seat_row`, `seat_column` LIMIT 1");
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
            'car_number' => (int) $carNumber,
            'seats' => $seatInformationList,
            'cars' => $simpleCarInformationList,
        ];

        return $response->withJson($carInformation);
    }

    public function trainReservationHandler(Request $request, Response $response, array $args)
    {
        /**
         * request payload
         *
         * string                     `json:"date"`
         * string                     `json:"train_name"`
         * string                     `json:"train_class"`
         * int                        `json:"car_number"`
         * bool                       `json:"is_smoking_seat"`
         * string                     `json:"seat_class"`
         * string                     `json:"departure"`
         * string                     `json:"arrival"`
         * int                        `json:"child"`
         * int                        `json:"adult"`
         * string                     `json:"Column"`
         * []{int row, string column} `json:"seats"`
          */
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withJson($this->errorResponse("JSON parseに失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        // 乗車日の日付表記統一
        $date = new \DateTime($payload['date']);
        if (! $this->checkAvailableDate($date)) {
            return $response->withJson($this->errorResponse("予約可能期間外です"), StatusCode::HTTP_NOT_FOUND);
        }

        $date = $date->setTimezone(new DateTimeZone("Asia/Tokyo"));

        $this->dbh->beginTransaction();
        // 止まらない駅の予約を取ろうとしていないかチェックする
        // 列車データを取得
        $stmt = $this->dbh->prepare("SELECT * FROM `train_master` WHERE `date`=? AND `train_class`=? AND `train_name`=?");
        $stmt->execute([
            $date->format(self::DATE_SQL_FORMAT),
            $payload['train_class'],
            $payload['train_name'],
        ]);
        $tmas = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($tmas === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("列車データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }

        // 列車自体の駅IDを求める
        $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
        $stmt->execute([$tmas['start_station']]);
        $departureStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($departureStation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("リクエストされた列車の始発駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }
        $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
        $stmt->execute([$tmas['last_station']]);
        $arrivalStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($arrivalStation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("リクエストされた列車の終着駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }

        // リクエストされた乗車区間の駅IDを求める
        $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
        $stmt->execute([$payload['departure']]);
        $fromStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($fromStation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("乗車駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }

        $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
        $stmt->execute([$payload['arrival']]);
        $toStation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($toStation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("降車駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }


        switch ($payload['train_class']) {
            case '最速':
                if (! (bool) $fromStation['is_stop_express'] || ! (bool) $toStation['is_stop_express']) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("最速の止まらない駅です"), StatusCode::HTTP_BAD_REQUEST);
                }
                break;
            case '中間':
                if (! (bool) $fromStation['is_stop_semi_express'] || ! (bool) $toStation['is_stop_semi_express']) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("中間の止まらない駅です"), StatusCode::HTTP_BAD_REQUEST);
                }
                break;
            case '遅いやつ':
                if (! (bool) $fromStation['is_stop_local'] || ! (bool) $toStation['is_stop_local']) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("遅いやつの止まらない駅です"), StatusCode::HTTP_BAD_REQUEST);
                }
                break;
            default:
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("リクエストされた列車クラスが不明です"), StatusCode::HTTP_BAD_REQUEST);
        }

        // 運行していない区間を予約していないかチェックする
        if ((bool) $tmas['is_nobori']) {
            if ($fromStation['id'] > $departureStation['id'] || $toStation['id'] > $departureStation['id']) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("リクエストされた区間に列車が運行していない区間が含まれています"), StatusCode::HTTP_BAD_REQUEST);
            }

            if ($arrivalStation['id'] >= $fromStation['id'] || $arrivalStation['id'] > $toStation['id']) {
                return $response->withJson($this->errorResponse("リクエストされた区間に列車が運行していない区間が含まれています"), StatusCode::HTTP_BAD_REQUEST);
            }
        } else {
            if ($fromStation['id'] < $departureStation['id'] || $toStation['id'] < $departureStation['id']) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("リクエストされた区間に列車が運行していない区間が含まれています"), StatusCode::HTTP_BAD_REQUEST);
            }

            if ($arrivalStation['id'] <= $fromStation['id'] || $arrivalStation['id'] < $toStation['id']) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("リクエストされた区間に列車が運行していない区間が含まれています"), StatusCode::HTTP_BAD_REQUEST);
            }
        }

        // あいまい座席検索 seatsが空白の時に発動する
        switch (count($payload['seats'])) {
            case 0:
                if ($payload['seat_class'] === 'non-reserved') {
                    // // non-reservedはそもそもあいまい検索もせずダミーのRow/Columnで予約を確定させる。
                    break;
                }
                // 当該列車・号車中の空き座席検索
                $stmt = $this->dbh->prepare("SELECT * FROM `train_master` WHERE `date`=? AND `train_class`=? AND `train_name`=?");
                $stmt->execute([
                    $date->format(self::DATE_SQL_FORMAT),
                    $payload['train_class'],
                    $payload['train_name'],
                ]);
                $train = $stmt->fetch(PDO::FETCH_ASSOC);
                if ($train === false) {
                    $this->dbh->rollBack();
                }

                $usableTrainClassList = $this->getUsableTrainClassList($fromStation, $toStation);
                $usable = in_array($train['train_class'], $usableTrainClassList);
                if (!$usable) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("invalid train_class"), StatusCode::HTTP_BAD_REQUEST);
                }

                for ($carnum = 1; $carnum <= 16; $carnum++) {
                    // 指定した車両内の座席のうち座席クラス等の条件に一致するもののみ抽出
                    $stmt = $this->dbh->prepare("SELECT * FROM `seat_master` WHERE `train_class`=? AND `car_number`=? AND `seat_class`=? AND `is_smoking_seat`=? ORDER BY `seat_row`, `seat_column`");
                    $stmt->execute([
                        $payload['train_class'],
                        $carnum,
                        $payload['seat_class'],
                        $payload['is_smoking_seat']
                    ]);
                    $seatList = $stmt->fetchAll(PDO::FETCH_ASSOC);
                    if ($seatList === false) {
                        $this->dbh->rollBack();
                        return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
                    }

                    if (count($seatList) == 0) {
                        // 条件を満たす座席がない車両だったので次へ
                        continue;
                    }

                    foreach ($seatList as $seat) {
                        $s = [
                            'row' => $seat['seat_row'],
                            'column' => $seat['seat_column'],
                            'class' => $seat['seat_class'],
                            'is_smoking_seat' => (bool) $seat['is_smoking_seat'],
                            'is_occupied' => false,
                        ];
                        $stmt = $this->dbh->prepare("SELECT s.* FROM `seat_reservations` s, `reservations` r WHERE r.`date` =? AND r.`train_class` =? AND r.`train_name` =? AND `car_number` =? AND `seat_row` =? AND `seat_column` =? FOR UPDATE");
                        $stmt->execute([
                            $date->format(self::DATE_SQL_FORMAT),
                            $seat['train_class'],
                            $payload['train_name'],
                            $seat['car_number'],
                            $seat['seat_row'],
                            $seat['seat_class'],
                        ]);
                        $seatReservationList = $stmt->fetchAll(PDO::FETCH_ASSOC);
                        if ($seatReservationList === false) {
                            $this->dbh->rollBack();
                            return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
                        }
                        foreach ($seatReservationList as $seatReservation) {
                            $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `reservation_id` =? FOR UPDATE");
                            $stmt->execute([
                                $seatReservation['reservation_id']
                            ]);
                            $reservation = $stmt->fetch(PDO::FETCH_ASSOC);
                            if ($reservation === false) {
                                $this->dbh->rollBack();
                                return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
                            }

                            $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
                            $departureStation = $stmt->execute([$reservation['departure']]);
                            if ($departureStation === false) {
                                $this->dbh->rollBack();
                                return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
                            }

                            $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
                            $arrivalStation = $stmt->execute([$reservation['arrival']]);
                            if ($arrivalStation === false) {
                                $this->dbh->rollBack();
                                return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_BAD_REQUEST);
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

                    // 曖昧予約席とその他の候補席を選出
                    $reserved = false;
                    $vargue = true;
                    $seatnum = ($payload['adult'] + $payload['child'] -1); // 予約する座席の合計数, 全体の人数からあいまい指定席分を引いておく
                    // A/B/C/D/Eを指定しなければ、空いている適当な指定席を取るあいまいモード
                    if ($payload['Column'] === "") {
                        $seatnum = ($payload['adult'] + $payload['child']); // あいまい指定せず大人＋小人分の座席を取る
                        $reserved = false; // dummy
                        $vargue = false;   // dummy
                    }

                    // シート分だけ回して予約できる席を検索
                    $i = 0;
                    $vagueSeat = [];
                    $candidateSeats = [];
                    foreach ($seatInformationList as $seat) {
                        if (($seat['column'] == $payload['Column']) && (! (bool) $seat['is_occupied']) && (! $reserved) && ($vargue)) {
                            $vagueSeat['row'] = $seat['row'];
                            $vagueSeat['column'] = $seat['column'];
                        } elseif ((! (bool) $seat['is_occupied']) && ($i < $seatnum)) {
                            $candidateSeats[] = [
                              'row' => $seat['row'],
                              'column'  => $seat['column'],
                            ];
                            $i++;
                        }
                    }

                    // あいまい席が見つかり、予約できそうだった
                    if ($vargue === true && $reserved === true) {
                        $payload['seats'][] = $vagueSeat;
                    }
                    //  候補席があった
                    if ($i > 0) {
                        foreach ($candidateSeats as $c) {
                            $payload['seats'][] = $c;
                        }
                    }

                    if (count($payload['seats']) < ($payload['adult'] = $payload['child'])) {
                        // リクエストに対して席数が足りてない
                        // 次の号車にうつしたい
                        // fmt.Println("-----------------")
                        // fmt.Printf("現在検索中の車両: %d号車, リクエスト座席数: %d, 予約できそうな座席数: %d, 不足数: %d\n", carnum, req.Adult+req.Child, len(req.Seats), req.Adult+req.Child-len(req.Seats))
                        // fmt.Println("リクエストに対して座席数が不足しているため、次の車両を検索します。")
                        if ($carnum === 16) {
                            // fmt.Println("この新幹線にまとめて予約できる席数がなかったから検索をやめるよ")
                            $payload['seats'] = [];
                            break;
                        }
                    }
                    // fmt.Printf("空き実績: %d号車 シート:%v 席数:%d\n", carnum, req.Seats, len(req.Seats))
                    if (count($payload['seats']) >= ($payload['adult'] + $payload['child'])) {
                        // fmt.Println("予約情報に追加したよ")
                        $payload['seats'] = array_slice($payload['seats'], 0, ($payload['adult'] + $payload['child']));
                        $payload['car_number'] = $carnum;
                        break;
                    }
                }
                if (count($payload['seats']) === 0) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("あいまい座席予約ができませんでした。指定した席、もしくは1車両内に希望の席数をご用意できませんでした。"), StatusCode::HTTP_BAD_REQUEST);
                }
                // no break
            default:
                // 座席情報のValidate
                foreach ($payload['seats'] as $z) {
                    $stmt = $this->dbh->prepare("SELECT * FROM `seat_master` WHERE `train_class` =? AND `car_number` =? AND `seat_column` =? AND `seat_row` =? AND `seat_class` =?");
                    $stmt->execute([
                        $payload['train_class'],
                        $payload['car_number'],
                        $z['column'],
                        $z['row'],
                        $payload['seat_class'],
                    ]);
                    $seatList = $stmt->fetch(PDO::FETCH_ASSOC);
                    if ($seatList === false) {
                        $this->dbh->rollBack();
                        return $response->withJson($this->errorResponse("リクエストされた座席情報は存在しません。号車・喫煙席・座席クラスなど組み合わせを見直してください"), StatusCode::HTTP_BAD_REQUEST);
                    }
                }
                break;
        }
        // 当該列車・列車名の予約一覧取得
        $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE date=? AND `train_class`=? AND `train_name`=? FOR UPDATE");
        $stmt->execute([
            $date->format(self::DATE_SQL_FORMAT),
            $payload['train_class'],
            $payload['train_name'],
        ]);
        $reservations = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($reservations === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("列車予約情報の取得に失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        foreach ($reservations as $reservation) {
            if ($payload['seat_class'] === 'non-reserved') {
                break;
            }
            // train_masterから列車情報を取得(上り・下りが分かる)
            $stmt = $this->dbh->prepare("SELECT * FROM `train_master` WHERE `date`=? AND `train_class`=? AND `train_name`=?");
            $stmt->execute([
                $date->format(self::DATE_SQL_FORMAT),
                $payload['train_class'],
                $payload['train_name'],
            ]);
            $tmas = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($tmas === false) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("列車データがみつかりません"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
            }

            // 予約情報の乗車区間の駅IDを求める
            // from
            $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
            $stmt->execute([$reservation['departure']]);
            $reservedFromStation = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($reservedFromStation === false) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("予約情報に記載された列車の乗車駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
            }

            // to
            $stmt = $this->dbh->prepare("SELECT * FROM `station_master` WHERE `name`=?");
            $stmt->execute([$reservation['arrival']]);
            $reservedToStation = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($reservedToStation === false) {
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("予約情報に記載された列車の降車駅データがみつかりません"), StatusCode::HTTP_NOT_FOUND);
            }

            // 予約の区間重複判定
            $secdup = false;
            if ((bool) $tmas['is_nobori'] === true) {
                if (($toStation['id'] < $reservedToStation['id']) && ($fromStation['id'] <= $reservedFromStation['id'])) {
                    // pass
                } elseif (($toStation['id'] >= $reservedToStation['id']) && ($fromStation['id'] > $reservedFromStation['id'])) {
                    // pass
                } else {
                    $secdup = true;
                }
            } else {
                if (($fromStation['id'] < $reservedFromStation['id']) && $toStation['id'] <= $reservedToStation['id']) {
                    // pass
                } elseif (($fromStation['id'] >= $reservedToStation['id']) && ($toStation['id'] > $reservedToStation['id'])) {
                    // pass
                } else {
                    $secdup = true;
                }
            }

            if ($secdup) {
                // 区間重複の場合は更に座席の重複をチェックする
                $stmt = $this->dbh->prepare("SELECT * FROM `seat_reservations` WHERE `reservation_id`=? FOR UPDATE");
                $stmt->execute([
                    $reservation['reservation_id'],
                ]);
                $seatReservations = $stmt->fetchAll(PDO::FETCH_ASSOC);
                if ($seatReservations === false) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("座席予約情報の取得に失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                }
                foreach ($seatReservations as $v) {
                    foreach ($payload['seats'] as $seat) {
                        if (($v['car_number'] == $payload['car_number']) && ($v['seat_row'] == $seat['row'] && ($v['seat_column'] == $seat['column']))) {
                            $this->dbh->rollBack();
                            return $response->withJson($this->errorResponse("リクエストに既に予約された席が含まれています"), StatusCode::HTTP_BAD_REQUEST);
                        }
                    }
                }
            }
        }
        // 3段階の予約前チェック終わり

        // 自由席は強制的にSeats情報をダミーにする（自由席なのに席指定予約は不可）
        if ($payload['seat_class'] === 'non-reserved') {
            $payload['seats'] = [];
            $payload['car_number'] = 0;
            for ($num=0; $num < ($payload['adult'] + $payload['child']); $num++) {
                $payload['seats'][] = [
                    'row' => 0,
                    'column' => "",
                ];
            }
        }

        // 運賃計算
        try {
            switch ($payload['seat_class']) {
                case 'premium':
                    $fare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $payload['train_class'], 'premium');
                    break;
                case 'reserved':
                    $fare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $payload['train_class'], 'reserved');
                    break;
                case 'non-reserved':
                    $fare = $this->fareCalc($date, $fromStation['id'], $toStation['id'], $payload['train_class'], 'non-reserved');
                    break;
                default:
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("リクエストされた座席クラスが不明です"), StatusCode::HTTP_BAD_REQUEST);
            }
            $sumFare = ($payload['adult'] * $fare) + (($payload['child'] * $fare) / 2);
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }

        // userID取得。ログインしてないと怒られる。
        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }

        // 予約ID発行と予約情報登録
        try {
            $stmt = $this->dbh->prepare("INSERT INTO `reservations` (`user_id`, `date`, `train_class`, `train_name`, `departure`, `arrival`, `status`, `payment_id`, `adult`, `child`, `amount`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
            $stmt->execute([
                $user['id'],
                $date->format(self::DATE_SQL_FORMAT),
                $payload['train_class'],
                $payload['train_name'],
                $payload['departure'],
                $payload['arrival'],
                "requesting",
                "a",
                $payload['adult'],
                $payload['child'],
                $sumFare,
            ]);
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("予約の保存に失敗しました"), StatusCode::HTTP_BAD_REQUEST);
        }
        $reservation_id = $this->dbh->lastInsertId();

        //席の予約情報登録
        //reservationsレコード1に対してseat_reservationstが1以上登録される
        foreach ($payload['seats'] as $v) {
            try {
                $stmt = $this->dbh->prepare("INSERT INTO `seat_reservations` (`reservation_id`, `car_number`, `seat_row`, `seat_column`) VALUES (?, ?, ?, ?)");
                $stmt->execute([
                    $reservation_id,
                    $payload['car_number'],
                    $v['row'],
                    $v['column']
                ]);
            } catch (\PDOException $e) {
                $this->logger->error($e->getMessage());
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("座席予約の登録に失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
            }
        }
        $this->dbh->commit();

        return $response->withJson([
            'reservation_id' => (int) $reservation_id,
            'amount' => $sumFare,
            'is_ok' => true,
        ], StatusCode::HTTP_OK);
    }

    public function reservationPaymentHandler(Request $request, Response $response, array $args)
    {
        /*
            支払い及び予約確定API
            POST /api/train/reservation/commit
            {
                "card_token": "161b2f8f-791b-4798-42a5-ca95339b852b",
                "reservation_id": "1"
            }

            前段でフロントがクレカ非保持化対応用のpayment-APIを叩き、card_tokenを手に入れている必要がある
            レスポンスは成功か否かのみ返す
        */
        /**
         * payload
         *
         * string `json:"card_token"`
         * int    `json:"reservation_id"`
         */
        $payload = $this->jsonPayload($request);
        $this->dbh->beginTransaction();
        // 予約IDで検索
        $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `reservation_id`=?");
        $stmt->execute([
            $payload['reservation_id'],
        ]);
        $reservation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($reservation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("予約情報がみつかりません"), StatusCode::HTTP_NOT_FOUND);
        }

        // 支払い前のユーザチェック。本人以外のユーザの予約を支払ったりキャンセルできてはいけない。
        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }

        if ($reservation['user_id'] !== $user['id']) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("他のユーザIDの支払いはできません"), StatusCode::HTTP_UNAUTHORIZED);
        }

        // 予約情報の支払いステータス確認
        switch ($reservation['status']) {
            case 'done':
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("既に支払いが完了している予約IDです"), StatusCode::HTTP_FORBIDDEN);
                break;
            default:
                break;
        }

        // 決済する
        $payInfo = [
            "payment_information" => [
            'card_token' => $payload['card_token'],
            'reservation_id' => $payload['reservation_id'],
            'amount' => $reservation['amount'],
            ]
        ];
        $payment_api = Environment::get('PAYMENT_API', 'http://payment:5000');
        $http_client = new Client();
        try {
            $r = $http_client->post($payment_api . '/payment', [
                'json' => $payInfo,
            ]);
        } catch (RequestException $e) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("HTTP POSTに失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        if ($r->getStatusCode() != StatusCode::HTTP_OK) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("決済に失敗しました。カードトークンや支払いIDが間違っている可能性があります"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        /**
         * string `json:"payment_id"`
         * bool   `json:"is_ok"`
         */
        $output = json_decode($r->getBody(), true);

        // 予約情報の更新
        try {
            $stmt = $this->dbh->prepare("UPDATE `reservations` SET `status`=?, `payment_id`=? WHERE `reservation_id`=?");
            $stmt->execute([
                "done",
                $output['payment_id'],
                $payload['reservation_id'],
            ]);
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("予約情報の更新に失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        $this->dbh->commit();
        return $response->withJson(['is_ok' => true], StatusCode::HTTP_OK);
    }

    public function userReservationsHandler(Request $request, Response $response, array $args)
    {
        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }

        $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `user_id`=?");
        $stmt->execute([$user['id']]);
        $reservationList = $stmt->fetchAll(PDO::FETCH_ASSOC);
        if ($reservationList === false) {
            return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_UNAUTHORIZED);
        }

        $res = [];
        try {
            foreach ($reservationList as $r) {
                $res[] = $this->makeReservationResponse($r);
            }
        } catch (\DomainException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }
        return $response->withJson($res, StatusCode::HTTP_OK);
    }

    public function userReservationResponseHandler(Request $request, Response $response, array $args)
    {
        $id = $args['id'] ?? 0;
        if ($id === 0) {
            return $response->withJson($this->errorResponse("incorrect item id"), StatusCode::HTTP_BAD_REQUEST);
        }

        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }

        $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `reservation_id`=? AND `user_id`=?");
        $stmt->execute([
            $id,
            $user['id']
        ]);
        $reservation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($reservation === false) {
            return $response->withJson($this->errorResponse("Reservation not found"), StatusCode::HTTP_NOT_FOUND);
        }
        try {
            $reservationResponse = $this->makeReservationResponse($reservation);
        } catch (\DomainException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_BAD_REQUEST);
        }

        return $response->withJson($reservationResponse, StatusCode::HTTP_OK);
    }

    public function userReservationCancelHandler(Request $request, Response $response, array $args)
    {
        $id = $args['id'] ?? 0;
        if ($id === 0) {
            return $response->withJson($this->errorResponse("incorrect item id"), StatusCode::HTTP_BAD_REQUEST);
        }

        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }

        $this->dbh->beginTransaction();
        $stmt = $this->dbh->prepare("SELECT * FROM `reservations` WHERE `reservation_id`=? AND `user_id`=?");
        $stmt->execute([
            $id,
            $user['id'],
        ]);
        $reservation = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($reservation === false) {
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse("reservations naiyo"), StatusCode::HTTP_BAD_REQUEST);
        }

        switch ($reservation['status']) {
            case 'rejected':
                $this->dbh->rollBack();
                return $response->withJson($this->errorResponse("何らかの理由により予約はRejected状態です"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
                break;
            case 'done':
                // 支払いをキャンセルする
                $payInfo = ['payment_id' => $reservation['payment_id']];
                $payment_api = Environment::get('PAYMENT_API', 'http://payment:5000');
                $http_client = new Client();
                try {
                    $r = $http_client->delete($payment_api . sprintf("/payment/%s", $reservation['payment_id']), [
                        'json' => $payInfo,
                        'timeout' => 10,
                    ]);
                } catch (RequestException $e) {
                    return $response->withJson($this->errorResponse("HTTP DELETEに失敗しました"), StatusCode::HTTP_BAD_REQUEST);
                }
                if ($r->getStatusCode() != StatusCode::HTTP_OK) {
                    $this->dbh->rollBack();
                    return $response->withJson($this->errorResponse("決済に失敗しました。支払いIDが間違っている可能性があります"), StatusCode::HTTP_BAD_REQUEST);
                }
                $output = json_decode($r->getBody());
                break;
            default:
                // pass(requesting状態のものはpayment_id無いので叩かない)
                break;
        }

        try {
            $stmt = $this->dbh->prepare("DELETE FROM `reservations` WHERE `reservation_id`=? AND `user_id`=?");
            $stmt->execute([
                $id,
                $user['id']
            ]);

            $stmt = $this->dbh->prepare("DELETE FROM `seat_reservations` WHERE `reservation_id`=?");
            $stmt->execute([$id]);
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            $this->dbh->rollBack();
            return $response->withJson($this->errorResponse($this->dbh->errorInfo()), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }
        $this->dbh->commit();

        # TYPO cancel
        return $response->withJson($this->messageResponse('cancell complete'), StatusCode::HTTP_OK);
    }

    public function getAuthHandler(Request $request, Response $response, array $args)
    {
        try {
            $user = $this->getUser();
        } catch (\DomainException|\PDOException $e) {
            return $response->withJson($this->errorResponse($e->getMessage()), StatusCode::HTTP_UNAUTHORIZED);
        }
        return $response->withJson(['email' => $user['email']], StatusCode::HTTP_OK);
    }

    public function signUpHandler(Request $request, Response $response, array $args)
    {
        /**
         * request payload
         *
         * string `json:"email"`
         * string `json:"password"`
         */
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withJson($this->errorResponse("JSON parseに失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        $salt = random_bytes(1024);
        $superSecurePassword = hash_pbkdf2('sha256', $payload['password'], $salt, 100, 256, true);
        $stmt = $this->dbh->prepare("INSERT INTO `users` (`email`, `salt`, `super_secure_password`) VALUES (?, ?, ?)");
        try {
            $stmt->execute([
                $payload['email'],
                $salt,
                $superSecurePassword,
            ]);
        } catch (\PDOException $e) {
            $this->logger->error("DB error", $this->dbh->errorInfo());
            return $response->withJson($this->errorResponse("user registration failed"), StatusCode::HTTP_BAD_REQUEST);
        }

        return $response->withJson($this->messageResponse("registration complete"), StatusCode::HTTP_OK);
    }

    public function loginHandler(Request $request, Response $response, array $args)
    {
        /**
         * request payload
         *
         * string `json:"email"`
         * string `json:"password"`
         */
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withJson($this->errorResponse("JSON parseに失敗しました"), StatusCode::HTTP_INTERNAL_SERVER_ERROR);
        }

        $stmt = $this->dbh->prepare("SELECT * FROM `users` WHERE `email`=?");
        $stmt->execute([$payload['email']]);
        $user = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($user === false) {
            return $response->withJson($this->errorResponse("authentication failed"), StatusCode::HTTP_FORBIDDEN);
        }

        $challengePassword = hash_pbkdf2('sha256', $payload['password'], $user['salt'], 100, 256, true);

        if ($user['super_secure_password'] !== $challengePassword) {
            return $response->withJson($this->errorResponse("authentication failed"), StatusCode::HTTP_FORBIDDEN);
        }

        $this->session->set('user_id', $user['id']);
        $this->session->set('csrf_token', bin2hex(random_bytes(10)));

        // TYPO authenticated
        return $response->withJson($this->messageResponse("autheticated"), StatusCode::HTTP_OK);
    }

    public function logoutHandler(Request $request, Response $response, array $args)
    {
        $this->session->set('user_id', 0);
        $this->session->set('csrf_token', bin2hex(random_bytes(10)));
        return $response->withJson($this->messageResponse("logged out"), StatusCode::HTTP_OK);
    }

    public function initialize(Request $request, Response $response, array $args)
    {
        $this->dbh->exec("TRUNCATE seat_reservations");
        $this->dbh->exec("TRUNCATE reservations");
        $this->dbh->exec("TRUNCATE users");

        return $response->withJson(["language" => "php", "available_days" => self::AVAILABLE_DAYS]);
    }

    public function settingsHandler(Request $request, Response $response, array $args)
    {
        return $response->withJson(["payment_api" => Environment::get('PAYMENT_API', 'http://localhost:5000')]);
    }
}
