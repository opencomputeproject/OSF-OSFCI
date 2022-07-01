$(document).ready ( function(){	
	console.log("Loader is loading")
	var uri = window.location.toString();
	if (uri.indexOf("?") != -1){
		var newuri = uri.substring(0, uri.indexOf("?"));
		window.history.replaceState({}, document.title, newuri);
	}
	var authrequest = $.get('/user/auth/authtoken', function(data){
		console.log(data)
		var jsonobj = JSON.parse(JSON.stringify(data))
		if (Object.hasOwn(jsonobj, 'Error')){
			console.log("Error")
			return
		}
		var myarray = Object.keys(jsonobj);
		for (let i = 0; i  < myarray.length; i++) {
			mylocalStorage[myarray[i]] = jsonobj[myarray[i]];
		}
		logged();
	}, "json");
	authrequest.fail(function(){
		console.log("Failed")
	});

});
