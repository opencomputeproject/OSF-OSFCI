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


function main(){


	clearDocument();
	loadHTML("html/main.html");

	var dropZoneiLo = document.getElementById('drop-zone-ilo');


	var startUploadiLo = function(files) {
	        console.log(files)
		var formData = new FormData();
		for(var i = 0; i < files.length; i++){
		    var file = files[i];
		    // Check the file type
		    console.log(file.type)
		    // Add the file to the form's data
		    formData.append('name', file.name);
		    formData.append('fichier', file);
  		}
		var xhr = new XMLHttpRequest();
		xhr.open('POST', window.location+'/ilofirmware', true);

		xhr.onload = function () {
				  if (xhr.status === 200) {
				    // File(s) uploaded
			 	    console.log('All good');
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
			console.log(percent);
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
                console.log(files)
                var formData = new FormData();
                for(var i = 0; i < files.length; i++){
                    var file = files[i];
                    // Check the file type
                    console.log(file.type)
                    // Add the file to the form's data
                    formData.append('name', file.name);
                    formData.append('fichier', file);
                }
                var xhr = new XMLHttpRequest();
                xhr.open('POST', window.location+'/biosfirmware', true);

                xhr.onload = function () {
                                  if (xhr.status === 200) {
                                    // File(s) uploaded
                                    console.log('All good');
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
			console.log("bios");
                        console.log(percent);
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
