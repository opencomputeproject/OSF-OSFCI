function managePasswordForgotten() {
	$('#passwordReset').click(function() { 
		// We must reset the document and send a link to the registered email
		// to a form where the end user can update the password
		clearDocument();
		loadHTML("html/navbar.html");
                loadJS("js/navbar.js");
		navbarHover();
                loginBtn();
		$('#dropdown').css("display","none");
		$(document.body).append("<center><h1>Please fill in the following form !</h1><center>");
		loadHTML("html/passwordForgotten.html");
		loadJS("js/forms.js");
                formSubmission('#passwordForgotten','generate_password_lnk_rst','Reset email successfully sent','Unknown user');
                loadHTML("footer.html");
	});
}
