'use strict';

app.controller('MainController', function($scope, $http, $interval) { 
  $scope.fetch = function() {
    $http.get('/status.json').success(function(data) {
      $scope.databases = data;
    });
  };

  $scope.fetch();
  $interval($scope.fetch, 1000);
});
