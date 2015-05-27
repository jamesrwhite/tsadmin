'use strict';

angular.module('tsadminFilters', []).filter('prettyUptime', function() {
  return function(time) {
    // Minutes
    if (time > 60 && time < 3600) {
      return Math.floor(time / 60) + 'm';
    // Hours
    } else if (time > 3600 && time < 86400) {
      var hours = Math.floor(time / 3600);
      var minutes = Math.floor(Math.floor(time % 3600) / 60);
      return hours + 'h ' + minutes + 'm';
    // Days
    } else if (time > 86400 && time < 604800) {
      var days = Math.floor(time / 86400);
      var hours = Math.floor(Math.floor(time % 86400) / 3600);
      return days + 'd ' + hours + 'h';
    // Weeks
    } else if (time > 604800 && time < 2419200) {
      var weeks = Math.floor(time / 604800);
      var days = Math.floor(Math.floor(time % 604800) / 86400);
      return weeks + 'w ' + days + 'd';
    // Months
    } else if (time > 2419200 && time < 31449600) {
      var months = Math.floor(time / 2419200);
      var weeks = Math.floor(Math.floor(time % 2419200) / 604800);
      return months + 'm ' + weeks + 'w';
    // Years!
    } else if (time > 31449600) {
      var years = Math.floor(time / 31449600);
      var months = Math.floor(Math.floor(time % 31449600) / 2419200);
      return years + 'y ' + months + 'm';
    // Seconds
    } else {
      return time + 's';
    }
  };
});
