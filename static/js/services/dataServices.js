'use strict'
var dataServices = angular.module('dataServices', ['ngResource']);

dataServices.factory('User', function($http, $resource, $cookieStore) {
    $http.defaults.headers.common['Authorization'] = $cookieStore.get('auth');
    
    return $resource('j/users/:id', {id:'@id'}, {
	get: { method: 'GET', url: 'j/users/show/:id' },
	update: { method: 'POST', url: 'j/users/update/:id' },
    })
});

dataServices.factory('Team', function($http, $resource, $cookieStore) {
    $http.defaults.headers.common['Authorization'] = $cookieStore.get('auth');
    console.log("Team data service");
    return $resource('j/teams/:id', {id:'@id', q:'@q'}, {
	get: { method: 'GET', url: 'j/teams/show/:id' },
	save: { method: 'POST', url: 'j/teams/new' },
	update: { method: 'POST', url: 'j/teams/update/:id' },
	delete: { method: 'POST', url: 'j/teams/destroy/:id' },
	search: { method: 'GET', url: 'j/teams/search?q=:q', isArray: true}
    })
});

dataServices.factory('Tournament', function($http, $resource, $cookieStore) {
    $http.defaults.headers.common['Authorization'] = $cookieStore.get('auth');
    
    return $resource('j/tournaments/:id', {id:'@id', q:'@q', teamId:'@teamId'}, {
	get: { method: 'GET', url: 'j/tournaments/show/:id' },
	save: { method: 'POST', url: 'j/tournaments/new' },
	update: { method: 'POST', url: 'j/tournaments/update/:id' },
	delete: { method: 'POST', url: 'j/tournaments/destroy/:id' },
	search: { method: 'GET', url: 'j/tournaments/search?q=:q', isArray: true},
	join: {method: 'POST', url: 'j/tournamentrels/create/:id'},
	leave: {method: 'POST', url: 'j/tournamentrels/destroy/:id'},
	joinAsTeam: {method: 'POST', url: 'j/tournamentteamrels/create/:id/:teamId'},
	leaveAsTeam: {method: 'POST', url: 'j/tournamentteamrels/destroy/:id/:teamId'},
	candidates: {method: 'GET', url: 'j/tournaments/candidates/:id', isArray: true}
    })
});

dataServices.factory('Invite', function($http, $log, $q, $cookieStore){
    $http.defaults.headers.common['Authorization'] = $cookieStore.get('auth');
    
    return {
	send: function(currentUser, emails){
	    var deferred = $q.defer();
	    $http({
		method: 'POST',
		url: '/j/invite',
		contentType: 'application/json',
		params:{ emails: emails, name: currentUser.Name } }).
		success(function(data,status,headers,config){
		    deferred.resolve(data);
		    $log.info(data, status, headers() ,config);
		}).
		error(function (data, status, headers, config){
		    $log.warn(data, status, headers(), config);
		    deferred.reject(status);
					});
	    return deferred.promise;
	}
    };
});
