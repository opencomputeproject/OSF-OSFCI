function navbarHover() {
var masterTimeout;
$('#dropdownMaster').on('mouseover', function(e) {
    var dropdown = $(e.target);
    var menu = $('#menuMaster');
    dropdown.addClass('show');
    menu.addClass('show');
});

$('#dropdownMaster').on('mouseout', function(e) {
    var dropdown = $(e.target);
    var menu = $('#menuMaster');
    setTimeout(function () {
    	if ( !($('#dropdownSecondary').is(':hover')) ) {
		if ( !($('#dropdownMaster').is(':hover')) ) {
		    	 dropdown.removeClass('show');
			 menu.removeClass('show');
			 $('#navbarDropdownMenuLink').removeClass('show');
		}
    	}
	
    }, 300);
});

$('#dropdownSecondary').on('mouseover', function(e) {
    clearTimeout(masterTimeout);
    var dropdown = $(e.target);
    var menu = $('#menuSecondary');
    dropdown.addClass('show');
    menu.addClass('show');
});

$('#dropdownSecondary').on('mouseout', function(e) {
    var dropdown = $(e.target);
    var menu = $('#menuSecondary');
    setTimeout(function () {
	if ( !($('#dropdownSecondary').is(':hover')) ) {
	    dropdown.removeClass('show');
	    menu.removeClass('show');
	}
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

$('#dl360').on('click', function(e) {
        InteractiveSession();
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

$('#features').on("click", function(event) {
        clearDocument();
        loadHTML("html/navbar.html");
        loadJS("js/navbar.js");
        navbarHover();
        loginBtn();
	loadHTML("html/features.html");
});
