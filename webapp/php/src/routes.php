<?php

use Slim\App;
use Slim\Http\Request;
use Slim\Http\Response;
use Slim\Http\StatusCode;

return function (App $app) {
    $container = $app->getContainer();

    // API
    $app->post('/initialize', \App\Service::class . ':initialize');
    $app->get('/api/settings', \App\Service::class . ':settingsHandler');

    // 予約関係
    $app->get("/api/stations", \App\Service::class . ':getStationsHandler');
    $app->get("/api/train/search", \App\Service::class . ':trainSearchHandler');
    $app->get("/api/train/seats", \App\Service::class . ':trainSeatsHandler');
    $app->post("/api/train/reserve", \App\Service::class . ':trainReservationHandler');
    $app->post("/api/train/reservation/commit", \App\Service::class . ':reservationPaymentHandler');

    // 認証関連
    $app->get("/api/train/reservation/commit", \App\Service::class . ':getAuthHandler');
    $app->post("/api/train/reservation/commit", \App\Service::class . ':signUpHandler');
    $app->post("/api/train/reservation/commit", \App\Service::class . ':loginHandler');
    $app->post("/api/train/reservation/commit", \App\Service::class . ':logoutHandler');
    //	mux.HandleFunc(pat.Get("/api/user/reservations"), userReservationsHandler)
//	mux.HandleFunc(pat.Get("/api/user/reservations/:item_id"), userReservationResponseHandler)
//	mux.HandleFunc(pat.Post("/api/user/reservations/:item_id/cancel"), userReservationCancelHandler)
};
