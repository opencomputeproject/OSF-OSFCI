
var mylocalStorage = {};
window.mylocalStorage = mylocalStorage;
var BMCUP=0;
var isPool=1;

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

var firmwarebmcuploaded=0;
var firmwarebiosuploaded=0;


function homebutton(){
	$('#btnbmc').on('click', function () {
		$.ajax({
    			type: "GET",
	                contentType: 'application/json',
    			url: window.location.origin + '/ci/startbmc/',
			success: function(response){
				console.log("Bmc em100 started");
				$('#bmcem100console').contents().find("head").remove();
                                $('#bmcem100console').contents().find("body").remove();
                                $('#bmcem100console').removeAttr("src");
				$('#bmcem100console').attr("src", window.location+"/console");
    			}
        	});
	});
	$('#btnsmbios').on('click', function () {
                $.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/startsmbios/',
                        success: function(response){
				$('#smbiosem100console').contents().find("head").remove();
                                $('#smbiosem100console').contents().find("body").remove();
                                $('#smbiosem100console').removeAttr("src");
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
                                console.log("Bmc console started");
                                $('#bmcconsole').attr("src", window.location+"/bmcconsole");
                        }
                });
        });
	$('#btnpoweroff').on('click', function () {
		BMCUP=0;		
                $('#bmcbutton').css("display","none");
		$('#bmcem100console').contents().find("head").remove();
                $('#bmcem100console').contents().find("body").remove();
		$('#bmcem100console').removeAttr("src");
		$('#smbiosem100console').contents().find("head").remove();
                $('#smbiosem100console').contents().find("body").remove();
                $('#smbiosem100console').removeAttr("src");
		$('#bmcconsole').contents().find("head").remove();
                $('#bmcconsole').contents().find("body").remove();
                $('#bmcconsole').removeAttr("src");
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

	// This function can start only if I can get a server otherwise 
	// it won't work

        clearDocument();
        loadHTML("html/navbar.html");
        loadJS("js/navbar.js");
        navbarHover();
        loginBtn();

	// We request a test node to the gateway
	// This request could be a little bit long
	loadHTML("html/wait.html");
        $.ajax({
                  type: "GET",
                  contentType: 'application/json',
                  url: window.location.origin + '/ci/'+ 'getServer',
                  success: function(response){
			var answer = JSON.parse(response);
			if ( answer.Waittime == "0" ) {
				$('#waitMessage').remove();
				run_ci(answer.Servername, parseInt(answer.RemainingTime));
			} else {
				console.log(response);
				// We must display a warning message
				loadHTML("html/waitserver.html");
				// We can run a countdown and we can restart the start_ci if 
				// the countdown arrive to 0
				// Set the date we're counting down to
				var secondWait = parseInt(answer.Waittime);
				// Update the count down every 1 second
			var x = setInterval(function() {
				var days = Math.floor(secondWait / ( 60 * 60 * 24));
				var hours = Math.floor((secondWait % (60 * 60 * 24)) / (60 * 60));
				var minutes = Math.floor((secondWait % ( 60 * 60)) / 60);
  				var seconds = Math.floor((secondWait % ( 60)) );

  				$("#countdown").html(days + "d " + hours + "h " + minutes + "m " + seconds + "s");
				$("#users").html(answer.Queue);
				secondWait = secondWait - 1;
				// If the count down is finished, write some text
  				if (secondWait < 0) {
				    // We stop the timer
				    clearInterval(x);
				  }
				}, 1000);
			}
                  }
        });
}

function run_ci(servername, RemainingSecond) {

	// We received a test node we can start the CI in interactive
	// and initiate a timer into the navbar ( 30 minutes )
	// When the timer is expired we close our CI session and move
	// To the next user or make a new request

	// We have to hide the various button from the navbar
	$('#loginNavbar').css("display","none");
	$('#input-navbar').css("display","none");
	$('#Home').css("display","none");
	$('#features').css("display","none");
	$('#help').css("display","none");
	$('#dropdown').css("display","none");

	$("#EndSession").css("display","");


	// We must determine if the controller node which has been allocated
	// is able to handle flash emulators as a pool or through a one by one 
	// approach. If this is a pool we can't offer the opportunity to load
	// more than once the uploaded bios or bmc images. If this is not a pool
	// we can reset each of the emulator and restart them

	$.ajax({
		type: "GET",
		contentType: 'application/json',
		url: window.location.origin + '/ci/isEmulatorsPool',
		success: function(response){
				console.log('Emulator Pool:'+response);
				var obj = jQuery.parseJSON(response);
				if ( obj.isPool === "no" ) {
					console.log("not a pool");
					isPool=0;
				}
			}
		});


	// The home button and most of the navbar button must be disabled

	var x = setInterval(function() {
                   var days = Math.floor(RemainingSecond / ( 60 * 60 * 24));
                   var hours = Math.floor((RemainingSecond % (60 * 60 * 24)) / (60 * 60));
                   var minutes = Math.floor((RemainingSecond % ( 60 * 60)) / 60);
                   var seconds = Math.floor((RemainingSecond % ( 60)) );

                   $("#counter").html(days + "d " + hours + "h " + minutes + "m " + seconds + "s");
                   RemainingSecond = RemainingSecond - 1;
                   // If the count down is finished, write some text

		   // Let's check if the BMC is up and running
		   // if yes we can activate the Go to BMC Web interface button !

		   if ( RemainingSecond % 60 == 0 && BMCUP == 0 ) {
			$.ajax({
                                type: "GET",
                                contentType: 'application/json',
                                url: window.location.origin + '/ci/bmcup',
                                success: function(response){
					if ( response == "\"1\"" ) {
						$('#bmcbutton').css("display","");
						$('#bmcbutton').on("click", function() {
							// we must redirect to the home page
							var win = window.open('https://'+window.location.hostname, '_blank');
							win.focus();
						});
						BMCUP=1;
					}
				}
			});	
		   }
		   if ( RemainingSecond < 300 ) {
			console.log("switching color");
			$('#counter').css('color', '#ff8c00');
		   }
		   if ( RemainingSecond < 60 ) {
			$('#counter').css("color", "#fb0000");
		   }
                   if (RemainingSecond < 0) {
                        // We stop the timer
                        clearInterval(x);
			// We have to reset the server and go back home !
	                $('#bmcem100console').contents().find("head").remove();
       		        $('#bmcem100console').contents().find("body").remove();
			$('#bmcem100console').removeAttr("src");
	                $('#smbiosem100console').contents().find("head").remove();
       		        $('#smbiosem100console').contents().find("body").remove();
	                $('#smbiosem100console').removeAttr("src");
	                $('#bmcconsole').contents().find("head").remove();
	                $('#bmcconsole').contents().find("body").remove();
	                $('#bmcconsole').removeAttr("src");
	                $.ajax({
	                        type: "GET",
	                        contentType: 'application/json',
	                        url: window.location.origin + '/ci/poweroff/',
	                        success: function(response){
	                                $.ajax({
	                                          type: "PUT",
	                                          contentType: 'application/json',
	                                          url: window.location.origin + '/ci/'+ 'stopServer/'+servername,
	                                          success: function(response){
	                                                // we move back to the main page
	                                                clearInterval(x);
	                                                $("#EndSession").css("display","none");
	                                                $("#modalSession").modal('hide');
							// We can be brutal here as we have timed out
							main();
	                                        }
	                                });
	                        }
	                });
                    }
                }, 1000);
	
        // We must also attach the end session confirmation button !

        $("#ConfirmSessionEnd").on("click", function() {
                // Ok if we come there we have to inform the server that
                // we want end our session. It must clean up the cache
                // and power off my machine
                // we can clean up my page and Display a thank you message
                $('#bmcem100console').contents().find("head").remove();
                $('#bmcem100console').contents().find("body").remove();
                $('#bmcem100console').removeAttr("src");
                $('#smbiosem100console').contents().find("head").remove();
                $('#smbiosem100console').contents().find("body").remove();
                $('#smbiosem100console').removeAttr("src");
                $('#bmcconsole').contents().find("head").remove();
                $('#bmcconsole').contents().find("body").remove();
                $('#bmcconsole').removeAttr("src");
		$("#ConfirmSessionEnd").css("display","none");
		$("#ConfirmSessionCancel").css("display","none");
		$('#modalHeader').html("");
		$('#modalBody').html(''
		+'	<div class="container-fluid">'
		+'	        <div class="row" style="margin-top:10px; margin-bottom:10px">'
		+'                       <div class="container-fluid">'
		+'                               <div class="row" style="margin-top:10px; margin-bottom:10px">'
		+'                                       <div class="col-sm">'
		+'                                       <div class="d-flex align-items-center">'
		+'                                        <strong>Please wait while we are cleaning your server ...</strong>'
		+'                                                 <div class="spinner-border ml-auto" role="status" aria-hidden="true"></div>'
		+'                                         </div>'
		+'                                       </div>'
		+'                              </div>'
		+'                       </div>'
		+'	        </div>'
		+'	</div>'
		);
                $.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/poweroff/',
                        success: function(response){
                                $.ajax({
                                          type: "PUT",
                                          contentType: 'application/json',
                                          url: window.location.origin + '/ci/'+ 'stopServer/'+servername,
                                          success: function(response){
                                                // we move back to the main page
						clearInterval(x);
						$("#EndSession").css("display","none");
						$("#modalSession").modal('hide');
						$('#modalSession').on('hidden.bs.modal', function (e) {
							main();
						});
                                        }
                                });
                        }
                });

        });

        loadHTML("html/main.html");

	$("#DownloadOpenBMC").on("click", function(){
                $("#modalDownloadBody").html("Downloading your openbmc image ...");
                $('#p-downloadbmc').css("display", "");
                $('#progress-downloadbmc').css("width","0%");
                $('#p-downloadlinuxboot').css("display", "none");
                Url_rel = '/user/'+mylocalStorage['username']+'/getOpenBMC';
                BuildSignedAuth(Url_rel, 'GET' , "application/octet-stream", function(authString) {
                $.ajax({
                         url: window.location.origin + Url_rel,
                         type: 'GET',
                         headers: {
                                  "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                                  "Content-Type" : "application/octet-stream",
                                  "myDate" : authString['formattedDate']
                                 },
                         contentType: 'application/octet-stream',
			 xhrFields:{
                           responseType: 'blob',
			   onprogress: function(progress)
                                {
                                    var percentage = Math.floor((progress.loaded / progress.total) * 100);
                                    $('#progress-downloadbmc').css("width",percentage+"%");
                                }
                         },
                         success: function(response) {
				$("#modalDownload").modal('hide');
				var fileName = "openbmc.rom" 
			 	var link=document.createElement('a');
				var url = window.URL || window.webkitURL;
				link.href=url.createObjectURL(response);
				link.download=fileName;
				link.click();
                               }
                       });
               });
        });
        $("#DownloadLinuxboot").on("click", function(){
		$("#modalDownloadBody").html("Downloading your linuxboot image ...");
                $('#p-downloadbmc').css("display", "none");
                $('#progress-downloadlinuxboot').css("width","0%");
                $('#p-downloadlinuxboot').css("display", "");
                Url_rel = '/user/'+mylocalStorage['username']+'/getLinuxBoot';
                BuildSignedAuth(Url_rel, 'GET' , "application/octet-stream", function(authString) {
                $.ajax({
                         url: window.location.origin + Url_rel,
                         type: 'GET',
                         headers: {
                                  "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                                  "Content-Type" : "application/octet-stream",
                                  "myDate" : authString['formattedDate']
                                 },
                         contentType: 'application/octet-stream',
			 xhrFields:{
                           responseType: 'blob',
			   onprogress: function(progress)
                                {
                                    var percentage = Math.floor((progress.loaded / progress.total) * 100);
                                    $('#progress-downloadlinuxboot').css("width",percentage+"%");
                                }
                         },
                         success: function(response) {
				$("#modalDownload").modal('hide');
				var fileName = "linuxboot.rom"
                                var link=document.createElement('a');
				var url = window.URL || window.webkitURL;
                                link.href=url.createObjectURL(response);
                                link.download=fileName;
                                link.click();
                               }
                       });
               });
        });

        $("#ConfirmDownload").on("click", function(){
                $("#modalDownload").modal('hide');
        });



        var dropZonebmc = document.getElementById('drop-zone-bmc');

	// We must get the O/S installer list from the server side
	$.ajax({
                        type: "GET",
                        contentType: 'application/json',
                        url: window.location.origin + '/ci/getosinstallers/',
                        success: function(response){
				if ( response != "" ) {
					console.log(response);
					// The list is not empty
					// We must parse the JSON content
					// the object contain an array of string
					var obj = JSON.parse(response)
					if ( obj.files.length > 0 ) {
						$('#osChoices').append('<div id="innerOSChoice" class="dropdown-menu" aria-labelledby="dropdownMenuButton"></div>');
						obj.files.forEach(function(value) {
							var singleWord;
							var tag;
							singleWord = value.split(".");
							tag = singleWord[0];
							if ( singleWord.len > 2 ) {
								for ( i = 1 ; i < (singleWord.length - 1 ) ; i++ ) {
									tag = tag + '_' + singleWord[i];
								}
							}
							$('#innerOSChoice').append('<a class="dropdown-item" id='+tag+'>'+value+'</a>');
							$('#'+tag).click(function(event){
								// we can push for a system uploading request
								console.log(value);
								$.ajax({
									type: "GET",
						                        contentType: 'application/json',
						                        url: window.location.origin + '/ci/getosinstallers/'+value,
						                        success: function(response){
										    // The console shall be up now
						                                    $('#osloader').attr("src", window.location+"/osloaderconsole");
										}
									});
							})
						});
					}
				}
                        }
        });

        var startUploadbmc = function(files) {
                var formData = new FormData();
                for(var i = 0; i < files.length; i++){
                    var file = files[i];
                    // Check the file type
                    // Add the file to the form's data
                    formData.append('name', file.name);
                    formData.append('fichier', file);
                }
                var xhr = new XMLHttpRequest();
                xhr.open('POST', window.location+'bmcfirmware/'+mylocalStorage['username'], true);

                xhr.onload = function () {
                                  if (xhr.status === 200) {
                                    // File(s) uploaded
                                    $('#bmcuploaded').show();
                                    $('#bmcfirmwarefeedback').html("<span class=\"badge alert-success pull-right\">Success</span>"+file.name);
				    $('#bmcem100console').contents().find("head").remove();
			            $('#bmcem100console').contents().find("body").remove();
                		    $('#bmcem100console').removeAttr("src");
                                    $('#bmcem100console').attr("src", window.location+"/console");
                                  } else {
                                    alert('Something went wrong uploading the file.');
                                  }
                             };
                xhr.upload.addEventListener('progress', function(e) {
                        var percent = e.loaded / e.total * 100;
                        $('#progress-bmc').css("width",Math.floor(percent)+"%");
                }, false);
                xhr.send(formData);
        }

        dropZonebmc.ondrop = function(e) {
                e.preventDefault();
		if ( isPool == 1 ) {
	                if ( firmwarebmcuploaded == 0 ) {
	                        this.className = 'upload-drop-zone';
	                        firmwarebmcuploaded =1;
	                        startUploadbmc(e.dataTransfer.files)
	                }
	                else
	                {
	                        alert('Only one firmware per session');
	                }
		}
		else
		{
			// We need to clean up the console
			$('#bmcem100console').contents().find("head").remove();
                        $('#bmcem100console').contents().find("body").remove();
                        $('#bmcem100console').removeAttr("src");
			// The connection to the console has been lost
			// We must inform the controller node that we want to get rid of the previous setup
			// and reset the relevant em100 emulator
			// before accepting the new file
			$.ajax({
		                type: "GET",
		                contentType: 'application/json',
		                url: window.location.origin + '/ci/resetEmulator/bmc',
		                success: function(response){
					console.log('BMC emulator has been reset');
                                }
                	});
			this.className = 'upload-drop-zone';
                        firmwarebmcuploaded =1;
                        startUploadbmc(e.dataTransfer.files)
		}
        }

        dropZonebmc.ondragover = function() {
                this.className = 'upload-drop-zone drop';
                return false;
        }

        dropZonebmc.ondragleave = function() {
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
                xhr.open('POST', window.location+'biosfirmware/'+mylocalStorage['username'], true);

                xhr.onload = function () {
                                  if (xhr.status === 200) {
                                    // File(s) uploaded
                                    $('#biosuploaded').show();
                                    $('#biosfirmwarefeedback').html("<span class=\"badge alert-success pull-right\">Success</span>"+file.name);
				    $('#smbiosem100console').contents().find("head").remove();
                                    $('#smbiosem100console').contents().find("body").remove();
                                    $('#smbiosem100console').removeAttr("src");
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
	$('#githubLinuxboot').on('click', function(e) {
		$('#githubLinuxboot').removeClass("text-danger");
		$('#githubLinuxboot').removeClass("is-invalid");
	})

	$('#btnbuildsmbios').on('click', function(e) {
		// We must put the value to the compile server as to kick a build
		// That request has to be signed and must be protected by the 
		// user credential as to avoid server side overload
		// Let's sort out the user input
		 input = $('#githubLinuxboot').val();
		 if ( input.trim().replace(/\s\s+/g, ' ').split(/\W/).length < 2 )
		 {
			$('#githubLinuxboot').addClass("text-danger is-invalid");
		 }
		 else
		 {
			 Data = input+' hpe/dl360gen10 1';
			 Url_rel = '/ci/buildbiosfirmware/'+mylocalStorage['username'];
			 BuildSignedAuth(Url_rel, 'PUT' , "text/plain", function(authString) {
			 $.ajax({
		         	 url: window.location.origin + Url_rel,
			         type: 'PUT',
				 headers: {
			              "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
			              "Content-Type" : "text/plain",
			              "myDate" : authString['formattedDate']
		                 },
			         data: Data,
			         contentType: 'text/plain',
			         success: function(response) {
					// The process to load the code is running
					// the response contain the code from the ttyd which has kicked off the build
					// We can allocate that code to the BIOS iframe and we shall be receiving build input
		                        $('#smbiosem100console').contents().find("head").remove();
       			                $('#smbiosem100console').contents().find("body").remove();
                                        $('#smbiosem100console').removeAttr("src");
       	                         	$('#smbiosem100console').attr("src", window.location+"/smbiosbuildconsole");
		        	 }
		        	 });
		          });
		   }
	});

	$('#githubopenbmc').on('click', function(e) {
                $('#githubopenbmc').removeClass("text-danger");
                $('#githubopenbmc').removeClass("is-invalid");
        })

	$('#btnbuildopenbmc').on('click', function(e) {
                // We must put the value to the compile server as to kick a build
                // That request has to be signed and must be protected by the
                // user credential as to avoid server side overload
	         input = $('#githubopenbmc').val();
                 if ( input.trim().replace(/\s\s+/g, ' ').split(/\W/).length < 2 )
                 {
                        $('#githubopenbmc').addClass("text-danger is-invalid");
                 }
                 else
                 {
	                 Data = input+' dl360poc 1';
       	         	 Url_rel = '/ci/buildbmcfirmware/'+mylocalStorage['username'];
	                 BuildSignedAuth(Url_rel, 'PUT' , "text/plain", function(authString) {
       		         $.ajax({
                	         url: window.location.origin + Url_rel,
                         	type: 'PUT',
                         	headers: {
	                              "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
	                              "Content-Type" : "text/plain",
	                              "myDate" : authString['formattedDate']
	                         },
	                         data: Data,
	                         contentType: 'text/plain',
	                         success: function(response) {
	                                // The process to load the code is running
	                                // the response contain the code from the ttyd which has kicked off the build
	                                // We can allocate that code to the BIOS iframe and we shall be receiving build input
	                                $('#bmcem100console').contents().find("head").remove();
	                                $('#bmcem100console').contents().find("body").remove();
                                        $('#bmcem100console').removeAttr("src");
	                                $('#bmcem100console').attr("src", window.location+"/bmcbuildconsole");
	                         }
	                         });
                        });
		}
        });

	$('#btnLoadbuiltsmbios').on('click', function(e) {
		Url_rel = '/ci/loadbuiltsmbios/'+mylocalStorage['username'];
		BuildSignedAuth(Url_rel, 'PUT' , "text/plain", function(authString) {
		$.ajax({
			url: window.location.origin + Url_rel,
                         type: 'PUT',
                         headers: {
                              "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                              "Content-Type" : "text/plain",
                              "myDate" : authString['formattedDate']
                         },
                         contentType: 'text/plain',
                         success: function(response) {
                                // The process to build the code is running
                                // the response contain the code from the ttyd which has kicked off the build
                                // We can allocate that code to the BIOS iframe and we shall be receiving build input
                                $('#smbiosem100console').contents().find("head").remove();
                                $('#smbiosem100console').contents().find("body").remove();
                                $('#smbiosem100console').removeAttr("src");
                                $('#smbiosem100console').attr("src", window.location+"/smbiosconsole");
                         }
                         });
                        });	
	});
        $('#btnLoadbuiltopenbmc').on('click', function(e) {
                Url_rel = '/ci/loadbuiltopenbmc/'+mylocalStorage['username'];
                BuildSignedAuth(Url_rel, 'PUT' , "text/plain", function(authString) {
                $.ajax({
                        url: window.location.origin + Url_rel,
                         type: 'PUT',
                         headers: {
                              "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                              "Content-Type" : "text/plain",
                              "myDate" : authString['formattedDate']
                         },
                         contentType: 'text/plain',
                         success: function(response) {
                                // The process to build the code is running
                                // the response contain the code from the ttyd which has kicked off the build
                                // We can allocate that code to the BIOS iframe and we shall be receiving build input
                                $('#bmcem100console').contents().find("head").remove();
                                $('#bmcem100console').contents().find("body").remove();
                                $('#bmcem100console').removeAttr("src");
                                $('#bmcem100console').attr("src", window.location+"/console");
                         }
                         });
                        });
        });


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

function InteractiveSession() {
	start_ci();
}

function BuildSignedAuth(uri, op, contentType, callback) {
	var returnObject = {};
	var currentDate = new Date;
        var formattedDate = currentDate.toGMTString().replace( /GMT/, '+0000');
	var stringToSign = op +'\n\n'+contentType+'\n'+formattedDate+'\n'+uri
	console.log(stringToSign)
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

function disconnect()
{
	delete mylocalStorage['accessKey'];
	delete mylocalStorage['secretKey'];
	delete mylocalStorage['username'];
	// Wait 5s and redirect to mainpage
	setTimeout(function () {
		mainpage();
    	}, 5000);
}

function mainpage(){
	clearDocument();
	// Must load the default home page
	loadHTML("html/navbar.html");
	loadJS("js/navbar.js");
	navbarHover();
	loginBtn();
	loadHTML("html/home.html");
	$('#background').css('background-image', 'url(images/landing.png)').fadeIn(3000);
        $(document).ready(function () {
                    $('#background').animate({ opacity: 1 }, { duration: 2000 });
                    $('#background').fadeIn(3000);
        });

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
		if ( getUrlParameter('resetPassword') == "1" )
                {
			clearDocument();
                        loadHTML("html/navbar.html");
                        loadJS("js/navbar.js");
                        navbarHover();
                        loginBtn();
                        $(document.body).append("<center><h1>Welcome Back !</h1><center>");
                        loadHTML("html/resetPassword.html");
			$('#dropdown').css("display","none");
			loadJS("js/login.js");
                        $('#username').val(getUrlParameter('username'));
                        $('#username').prop('disabled', true);
                        $('#validation').val(getUrlParameter('validation'));
                        $('#validation').prop('disabled', true);
                        loadJS("js/forms.js");
                        formSubmission('#resetPassword','resetPassword','Password successfully reset','Reset link expired');
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
			$('#background').css('background-image', 'url(images/landing.png)').fadeIn(3000);
	      	        $(document).ready(function () {
	       	             $('#background').animate({ opacity: 1 }, { duration: 2000 });
	       	             $('#background').fadeIn(3000);
	       	         });
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
}

if ( getUrlParameter('loginValidated') == "1" )
{
	main();
}
