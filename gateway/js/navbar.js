var dropdownCodeToInsert = `
    <a class="nav-link dropdown-toggle" id="navbarDropdownMenuLink" data-toggle="dropdownMaster" 
       aria-haspopup="true" aria-expanded="false">Dropdown</a>
    <ul class="dropdown-menu dropdown-primary" aria-labelledby="navbarDropdownMenuLink" id="menuMaster">
      <li><a class="dropdown-item" id='MyAccount'>My Account</a></li>
      <li class="dropdown-submenu" id="dropdownSecondary">
        <a class="dropdown-toggle" data-toggle="dropdownSecondary" aria-haspopup="true" aria-expanded="false" id='InteractiveSession'>
          <span class="nav-label">Interactive Session</span><span class="caret"></span>
        </a>
        <ul class="dropdown-menu" aria-labelledby="InteractiveSession" id="menuSecondary">
          <li><a class="dropdown-item" id="dl360">dl360</a></li>
        </ul>
      </li>
    </ul>
  `;

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
				loginBtn();
		       		$(document.body).append("<center><h1>Welcome Back !</h1><center>");
			       	loadHTML("html/loginForm.html");
		       		loadJS("js/login.js");
		        	managePasswordForgotten();
		        	loadJS("js/forms.js");
		        	formSubmission('#login','get_token','','Password missmatch');
		        	loadHTML("footer.html");
		}
	});


	// We must check if we are logged in or not ?
	// and replace the button text
	if ( typeof(mylocalStorage) !== 'undefined' ) {
		if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
		{
			// we must change the login button by a Disconnect button
			$('#loginNavbar').html('Logout');
			// The navBar title must be the login name
			$("#dropdownMaster").append(dropdownCodeToInsert);
			$('#navbarDropdownMenuLink').html(mylocalStorage['username']);
			navbarHover()
			$('#MyAccount').on('click', function(e) {
				myAccount();
			});
			get_server_models_for_dropdown()
		}
		else {
			console.log("Error with html text insertion");
		}
	}	
		
}

$("#Home").on("click", function(event) {
	mainpage();
});

$("#download_key_button").click(function() {
	window.location.href = "./get_private_key";
});

$("#ack_button").click(function() {
	let nameString = "priv_key_ack"
	if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
	{
		nameString = nameString + "_username_" + mylocalStorage['username']
	}
    createCookie(nameString, 1, 30);
});


// tooltip element will appear/hide once the helptoggle is clicked
function enableDisableToolTip() {
    let toolTip = document.getElementsByClassName('tooltiptext')
    for (var i = 0; i < toolTip.length; i++) {
      if (typeof toolTip[i] === "undefined") {
        return
      }
      else if (toolTip[i].style.display == 'none') {
        toolTip[i].style.display = ''
      } else {
        toolTip[i].style.display = 'none'
      }
    }
  }

function popUp() {
	$('#myModal').modal('show');	
}

// We have to build the navbar production option
function get_server_models_for_dropdown() {
	$.ajax({
		type: "GET",
		contentType: 'application/json',
		url: window.location.origin + '/ci/get_server_models/',
		success: function(response){
				var obj = JSON.parse(response);
				var htmlcode = "";
				obj.forEach(function(item) {
					htmlcode = htmlcode + '<li><a class="dropdown-item" id="'+item.Product+'">' + item.Product + '</a></li>';
				});
				$('#menuSecondary').html(htmlcode);
				obj.forEach(function(item) {
					$('#'+item.Product).on('click', function(e) {
						InteractiveSession(item.Product);
					});
				});
		}
	});
}


