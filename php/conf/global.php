<?php

if (isset($_REQUEST['profile']) || isset($_ENV['PROFILE_ENABLED']) || !empty($_COOKIE['_profile'])) {
    try {
        include '/home/dev/.composer/vendor/autoload.php';

        if (class_exists('\Xhgui\Profiler\Profiler')) {
            $config = [
                'profiler' => \Xhgui\Profiler\Profiler::PROFILER_TIDEWAYS_XHPROF,
                'save.handler' => \Xhgui\Profiler\Profiler::SAVER_MONGODB,
                'save.handler.mongodb' => array(
                    'dsn' => 'mongodb://mongo:27017',
                    'database' => 'xhprof',
                    'options' => array(),
                    'driverOptions' => array(),
                ),
                'profiler.enable' => function () {
                    return true;
                },
            ];
            $profiler = new \Xhgui\Profiler\Profiler($config);
            $profiler->start();
        }
    } catch (\Throwable $t) {
    }
}
