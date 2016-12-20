angular.module('qmsk.e2', [
        'qmsk.e2.source',
        'qmsk.e2.web',
        'ngResource',
        'ngRoute',
        'ngWebSocket',
        'luegg.directives',
        'jsonFormatter',
])

.config(function($routeProvider) {
    $routeProvider
        .when('/main', {
            templateUrl: '/static/qmsk.e2/server/main.html',
            controller: 'MainCtrl',
            reloadOnSearch: false,
        })
        .when('/sources', {
            templateUrl: '/static/qmsk.e2/server/sources.html',
            controller: 'SourcesCtrl',
        })
        .when('/screens', {
            templateUrl: '/static/qmsk.e2/server/screens.html',
            controller: 'ScreensCtrl',
        })
        .when('/auxes', {
            templateUrl: '/static/qmsk.e2/server/auxes.html',
            controller: 'AuxesCtrl',
        })
        .when('/presets', {
            templateUrl: '/static/qmsk.e2/server/presets.html',
            controller: 'PresetsCtrl',
        })
        .when('/system', {
            templateUrl: '/static/qmsk.e2/server/system.html',
            controller: 'SystemCtrl',
        })
        .otherwise({
            redirectTo: '/main',
        });
})

.factory('Status', function($http) {
    var Status = {};

    $http.get('/api/status').then(
        function success(r) {
            Status.server = r.data.server;
            Status.mode = r.data.mode;
        }
    );

    return Status;
})

.factory('Index', function($http) {
    return function() {
        return $http.get('/api/').then(
            function success(r) {
                return r.data;
            }
        );
    };
})

.factory('Screen', function($resource) {
    return $resource('/api/screens/:id', { }, {
        get: {
            method: 'GET',
        },
        query: {
            method: 'GET',
            isArray: false,
        }
    }, {stripTrailingSlashes: true});
})

.factory('Aux', function($resource) {
    return $resource('/api/auxes/:id', { }, {
        get: {
            method: 'GET',
        },
        query: {
            method: 'GET',
            isArray: false,
        }
    }, {stripTrailingSlashes: true});
})

.factory('Preset', function($resource) {
    return $resource('/api/presets/:id', { }, {
        get: {
            method: 'GET',
            url: '/api/presets',
        },
        all: {
            method: 'GET',
            isArray: true,
        },
        query: {
            method: 'GET',
            isArray: false,
        }
    }, {stripTrailingSlashes: false});
})

.filter('dimensions', function() {
    return function(dimensions) {
        if (dimensions && dimensions.width && dimensions.height) {
            return dimensions.width + "x" + dimensions.height;
        } else {
            return null;
        }
    };
})

.controller('HeaderCtrl', function($scope, $location, Status, httpState) {
    $scope.status = Status;
    $scope.state = httpState;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };

})

.controller('StatusCtrl', function($scope, Events) {
    $scope.events = Events;
})

.controller('SystemCtrl', function($scope, Events) {
    $scope.events = Events;
})

.controller('MainCtrl', function($scope, $location, Index, $interval) {
    $scope.busy = false;
    $scope.error = null;

    $scope.reload = function() {
        if ($scope.busy) {
            return;
        } else {
            $scope.busy = true;
        }

        Index().then(
            function success(index) {
                $scope.busy = false;
                $scope.error = null;

                $scope.screens = index.screens
                $scope.sources = $.map(index.sources, function(source, id){
                    return source;
                });
            },
            function error(err) {
                console.log("MainCtrl: Index Error: " + err);

                $scope.busy = false;
                $scope.error = err;
            }
        );
    };

    $scope.selectOrder = function(order) {
        $scope.order = order;
        $scope.orderBy = function(){
            switch (order) {
            case 'source':
                return ['-type', 'name'];
            case 'preview':
                return ['preview_screens', 'program_screens'];
            case 'program':
                return ['program_screens', 'preview_screens'];
            default:
                return [];
            }
        }();

        $location.search('order', order || null);
    };
    $scope.selectOrder($location.search().order || 'source');

    $scope.reload();

    $scope.$on('qmsk.e2.event', function($e, event){
        // dumb :)
        $scope.reload();
    });
})

.controller('SourcesCtrl', function($scope, Source) {
    $scope.sources = Source.query();
})

.controller('ScreensCtrl', function($scope, Screen) {
    $scope.screens = Screen.query();
})

.controller('AuxesCtrl', function($scope, Aux) {
    $scope.auxes = Aux.query();
})

.controller('PresetsCtrl', function($scope, Preset, Screen, Aux) {
    $scope.collapseGroups = { };

    $scope.screens = Screen.query();
    $scope.auxes = Aux.query();
    
    // group
    $scope.presets = Preset.all(function (presets) {
        var groups = { };

        $.each(presets, function(id, preset) {
            // XXX: this is broken for non-array query()
            if (id[0] == '$' || !preset.group) {
                return;
            }

            var group = groups[preset.group];

            if (!group) {
                group = groups[preset.group] = [];
            }

            group.push(preset);
        });

        $scope.groups = $.map(groups, function(presets, id){
            return {id: id, presets: presets};
        });
    });
})

;
