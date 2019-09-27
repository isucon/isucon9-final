<?php

use Slim\App;
use Slim\Http\Request;
use Slim\Http\Response;
use Slim\Http\StatusCode;

return function (App $app) {
    $container = $app->getContainer();

    // API
    $app->post('/initialize', \App\Service::class . ':initialize');

};
