function navbarHover() {
$('body').on('mouseover mouseout', '.dropdown', function(e) {
    var dropdown = $(e.target).closest('.dropdown');
    var menu = $('.dropdown-menu', dropdown);
    dropdown.addClass('show');
    menu.addClass('show');
    setTimeout(function () {
        dropdown[dropdown.is(':hover') ? 'addClass' : 'removeClass']('show');
        menu[dropdown.is(':hover') ? 'addClass' : 'removeClass']('show');
    }, 300);
});
}

function loginBtn() {
$('#loginNavbar').on('click', function(e) {
	if ( typeof(mylocalStorage) !== 'undefined' ) 
	if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
	{
		disconnect();
	}
	else
	{
		clearDocument();
		loadHTML("html/navbar.html");
		loadJS("js/navbar.js");
		navbarHover();
		loginBtn();
       		$(document.body).append("<center><h1>Welcome Back !</h1><center>");
	       	loadHTML("html/loginForm.html");
       		loadJS("js/login.js");
        	managePasswordForgotten();
        	loadJS("js/forms.js");
        	formSubmission('#login','getToken','','Password missmatch');
        	loadHTML("footer.html");
	}
});

$('#MyAccount').on('click', function(e) {
	myAccount();
});

$('#MyProjects').on('click', function(e) {
        myProjects();
});
	// We must check if we are logged in or not ?
	// and replace the button text
	if ( typeof(mylocalStorage) !== 'undefined' )
	if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
	{
	        // we must change the login button by a Disconnect button
	        $('#loginNavbar').html('Logout');
		$('#navbarDropdownMenuLink').show();
		// The navBar title must be the login name
		$('#navbarDropdownMenuLink').html(mylocalStorage['username']);
	}
	else
		$('#navbarDropdownMenuLink').hide();
}

$("#Home").on("click", function(event) {
	mainpage();
});
