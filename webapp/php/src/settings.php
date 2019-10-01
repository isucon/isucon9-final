<?php

use App\Environment;

return [
    'settings' => [
        'displayErrorDetails' => true, // set to false in production
        'addContentLengthHeader' => false, // Allow the web server to send the content-length header
        'determineRouteBeforeAppMiddleware' => true,

        // Monolog settings
        'logger' => [
            'name' => 'isutrain',
            // 'path' => __DIR__ . '/../logs/app.log',
            'path' => 'php://stdout',
            'level' => \Monolog\Logger::DEBUG,
        ],

        // Database settings
        'database' => [
            'host' => Environment::get('MYSQL_HOSTNAME', '127.0.0.1'),
            'port' => Environment::get('MYSQL_PORT', '3306'),
            'username' => Environment::get('MYSQL_USER', 'isutrain'),
            'password' => Environment::get('MYSQL_PASSWORD', 'isutrain'),
            'dbname' => Environment::get('MYSQL_DATABASE', 'isutrain'),
        ],
        'app' => [
            'base_dir' => __DIR__ . '/../',
        ],
    ],
];
