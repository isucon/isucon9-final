<?php

use Slim\App;

return function (App $app) {
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
    $app->get("/api/auth", \App\Service::class . ':getAuthHandler');
    $app->post("/api/auth/signup", \App\Service::class . ':signUpHandler');
    $app->post("/api/auth/login", \App\Service::class . ':loginHandler');
    $app->post("/api/auth/logout", \App\Service::class . ':logoutHandler');
    $app->get("/api/user/reservations", \App\Service::class . ':userReservationsHandler');
    $app->get("/api/user/reservations/{id:\d+}", \App\Service::class . ':userReservationResponseHandler');
    $app->post("/api/user/reservations/{id:\d+}/cancel", \App\Service::class . ':userReservationCancelHandler');
};
