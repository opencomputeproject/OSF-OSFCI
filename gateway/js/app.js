
var mylocalStorage = {};
window.mylocalStorage = mylocalStorage;

function clearDocument(){
	$(document.body).empty();
}

function loadHTML(filename){
	jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
                $(document.body).append(data);
        });
        jQuery.ajaxSetup({async:true});
}

function getHTML(filename){
        jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
        	jQuery.ajaxSetup({async:true});
		return(data);
        });
}

function loadCSS(filename){
	jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
	$("<style>").prop("type", "text/css").html(data).appendTo("head");
        });
        jQuery.ajaxSetup({async:true});
}

function loadJS(filename){
        jQuery.ajaxSetup({async:false});
        jQuery.getScript(filename, function(data, textStatus, jqxhr) {
                });
        jQuery.ajaxSetup({async:true});
}

var firmwareilouploaded=0;
var firmwarebiosuploaded=0;


function homebutton(){
	$('#btnilo').on('click', function () {
		$.ajax({
    			type: "GET",
	                contentType: 'application/json',
    			url: window.location.origin + '/ci/startilo/',
			success: function(response){
				console.log("Ilo em100 started");
				$('#iloem100console').attr("src", window.location+"/console");
    			}
        	});
	});
	$('#btnsmbios').on('click', function () {
                $.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/startsmbios/',
                        success: function(response){
                                $('#smbiosem100console').attr("src", window.location+"/smbiosconsole");
                        }
                });
        });
	$('#btnpoweron').on('click', function () {
                $.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/poweron/',
                        success: function(response){
                                console.log("Ilo console started");
                                $('#iloconsole').attr("src", window.location+"/iloconsole");
                        }
                });
        });
	$('#btnpoweroff').on('click', function () {
		$('#iloem100console').removeAttr("src");
                $('#smbiosem100console').removeAttr("src");
                $('#iloconsole').removeAttr("src");
                $.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/poweroff/',
                        success: function(response){
                                console.log("System stopped");
                        }
                });
        });
}

function start_ci() {
        clearDocument();
        loadHTML("html/main.html");
        var dropZoneiLo = document.getElementById('drop-zone-ilo');


        var startUploadiLo = function(files) {
                var formData = new FormData();
                for(var i = 0; i < files.length; i++){
                    var file = files[i];
                    // Check the file type
                    // Add the file to the form's data
                    formData.append('name', file.name);
                    formData.append('fichier', file);
                }
                var xhr = new XMLHttpRequest();
                xhr.open('POST', window.location+'/ilofirmware', true);

                xhr.onload = function () {
                                  if (xhr.status === 200) {
                                    // File(s) uploaded
                                    $('#ilouploaded').show();
                                    $('#ilofirmwarefeedback').html("<span class=\"badge alert-success pull-right\">Success</span>"+file.name);
                                    $('#iloem100console').attr("src", window.location+"/console");
                                  } else {
                                    alert('Something went wrong uploading the file.');
                                  }
                             };
                xhr.upload.addEventListener('progress', function(e) {
                        var percent = e.loaded / e.total * 100;
                        $('#progress-ilo').css("width",Math.floor(percent)+"%");
                }, false);
                xhr.send(formData);
        }

        dropZoneiLo.ondrop = function(e) {
                e.preventDefault();
                if ( firmwareilouploaded == 0 ) {
                        this.className = 'upload-drop-zone';
                        firmwareilouploaded =1;
                        startUploadiLo(e.dataTransfer.files)
                }
                else
                {
                        alert('Only one firmware per session');
                }
        }

        dropZoneiLo.ondragover = function() {
                this.className = 'upload-drop-zone drop';
                return false;
        }

        dropZoneiLo.ondragleave = function() {
                this.className = 'upload-drop-zone';
                return false;
        }

        var dropZonebios = document.getElementById('drop-zone-bios');


        var startUploadbios = function(files) {
                var formData = new FormData();
                for(var i = 0; i < files.length; i++){
                    var file = files[i];
                    // Check the file type
                    // Add the file to the form's data
                    formData.append('name', file.name);
                    formData.append('fichier', file);
                }
                var xhr = new XMLHttpRequest();
                xhr.open('POST', window.location+'/biosfirmware', true);

                xhr.onload = function () {
                                  if (xhr.status === 200) {
                                    // File(s) uploaded
                                    $('#biosuploaded').show();
                                    $('#biosfirmwarefeedback').html("<span class=\"badge alert-success pull-right\">Success</span>"+file.name);
                                    $('#smbiosem100console').attr("src", window.location+"/smbiosconsole");
                                  } else {
                                    alert('Something went wrong uploading the file.');
                                  }
                             };
                xhr.upload.addEventListener('progress', function(e) {
                        var percent = e.loaded / e.total * 100;
                        $('#progress-bios').css("width",Math.floor(percent)+"%");
                }, false);
                xhr.send(formData);
        }

        dropZonebios.ondrop = function(e) {
                e.preventDefault();
                this.className = 'upload-drop-zone';
                // Only if a file was not uploaded
                if ( firmwarebiosuploaded == 0 ) {
                        this.className = 'upload-drop-zone';
                        firmwarebiosuploaded=1;
                        startUploadbios(e.dataTransfer.files)
                }
                else
                {
                        alert('Only one firmware per session');
                }
        }

        dropZonebios.ondragover = function() {
                this.className = 'upload-drop-zone drop';
                return false;
        }

        dropZonebios.ondragleave = function() {
                this.className = 'upload-drop-zone';
                return false;
        }



        homebutton();
}

var getUrlParameter = function getUrlParameter(sParam) {
    var sPageURL = window.location.search.substring(1),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : decodeURIComponent(sParameterName[1]);
        }
    }
};

function BuildSignedAuth(uri, op, contentType, callback) {
	var returnObject = {};
	var currentDate = new Date;
        var formattedDate = currentDate.toGMTString().replace( /GMT/, '+0000');
	var stringToSign = op +'\n\n'+contentType+'\n'+formattedDate+'\n'+uri
	returnObject['formattedDate'] = formattedDate;
        const buffer = new TextEncoder( 'utf-8' ).encode( stringToSign );
	if ( mylocalStorage['secretKey'] !== undefined && mylocalStorage['secretKey'].length > 0)
	{

		var hash = CryptoJS.HmacSHA1(stringToSign, mylocalStorage['secretKey']);
		returnObject['signedString'] = CryptoJS.enc.Base64.stringify(hash);
	}
	else
		returnObject['signedString'] = '';
	callback(returnObject);
}

function myAccount()
{
	clearDocument();
	loadHTML("html/navbar.html");
        loadJS("js/navbar.js");
	navbarHover();
	loginBtn();

	// We must put in place the layout here and allow various entries to be available
	// The first one is personal settings but others might be coming up
	var layout = '<div class="container-fluid"><div class="row" id="Row1">\
			<div class="col" style="width:10%" id="col0"></div>\
			<div closs="col" style="width:60%" id="col1"></div>\
			<div class="col" style="width:10%" id="col2"></div></div>\
			<div class="row"><div class="col" style="width:100%" id="col3"></div></div>';
        $(document.body).append(layout);

	loadJS("js/myaccount.js");
}

function logged()
{
	mainpage();
}

function mainpage(){
	clearDocument();
	// Must load the default home page
	loadHTML("html/navbar.html");
	loadJS("js/navbar.js");
	navbarHover();
	loginBtn();
	loadHTML("html/home.html");

	if (( "string" !== typeof(mylocalStorage['secretKey']) ) & ( "string" !== typeof(mylocalStorage['accessKey']) ))
	{
		$('#signup').css("display", "");
	}

//	loadJS("js/projects.js");
	loadJS("js/forms.js");
	loadJS("js/base.js");
	loadHTML("footer.html");
	formSubmission('#signup','createUser','User created - Please check your email','User exist');
}

function main(){
	if ( getUrlParameter('loginValidated') == "1" )
	{
		// We must check if the registration is ok
		clearDocument();
		loadHTML("html/navbar.html");
		loadJS("js/navbar.js");
		navbarHover();
		loginBtn();
                $(document.body).append("<center><h1>Welcome Back !</h1></center>");
		loadHTML("html/loginForm.html");
		loadJS("js/login.js");
		managePasswordForgotten();
		loadJS("js/forms.js");
		formSubmission('#login','getToken','','Password missmatch');
		loadHTML("html/footer.html");
	}
	else
	{
		clearDocument();
		loadHTML("html/navbar.html");
		loadJS("js/navbar.js");
		navbarHover();
		loginBtn();
		loadHTML("html/home.html");
		if (( "string" !== typeof(mylocalStorage['secretKey']) ) & ( "string" !== typeof(mylocalStorage['accessKey']) ))
		{
			$('#signup').css("display", "");
		}
		loadJS("js/forms.js");
		loadJS("js/base.js");
		loadHTML("html/footer.html");
		formSubmission('#signup','createUser','User created - Please check your email','User exist');
	}
}

if ( getUrlParameter('loginValidated') == "1" )
{
	main();
}
