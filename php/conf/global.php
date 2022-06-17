<?php

if (isset($_REQUEST['profile']) || isset($_ENV['PROFILE_ENABLED']) || !empty($_COOKIE['_profile'])) {
    try {
        include '/home/dev/.composer/vendor/autoload.php';

        if (class_exists('\Xhgui\Profiler\Profiler')) {
            $config = [
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

if (!function_exists('pdd')) {
    function pdd($data)
    {
        ini_set("highlight.comment", "#969896; font-style: italic");
        ini_set("highlight.default", "#FFFFFF");
        ini_set("highlight.html", "#D16568");
        ini_set("highlight.keyword", "#7FA3BC; font-weight: bold");
        ini_set("highlight.string", "#F2C47E");
        foreach (func_get_args() as $arg) {
            $output = highlight_string("<?php\n\n" . var_export($arg, true), true);
            echo "<div style=\"background-color: #1C1E21; padding: 1rem\">{$output}</div>";
        }
        die();
    }
}