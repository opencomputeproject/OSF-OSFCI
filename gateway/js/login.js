function managePasswordForgotten() {
	$('#passwordReset').click(function() { 
		// We must reset the document and send a link to the registered email
		// to a form where the end user can update the password
		clearDocument();
		loadHTML("html/navbar.html");
		navbarHover();
		$(document.body).append("<center><h1>Please fill in the following form !</h1><center>");
		loadHTML("html/passwordForgotten.html");
		loadJS("js/forms.js");
                formSubmission('#passwordForgotten','generatePasswordLnkRst','Reset email successfully sent','Unknown user');
                loadHTML("footer.html");
	});
}
