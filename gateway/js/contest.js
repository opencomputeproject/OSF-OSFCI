$(document).ready(function() {
	console.log(localStorage.getItem("username"))
	$('#btnstandardtests').on('click', function () {
		$.ajax({
			type: "GET",
			contentType: 'application/json',
			url: window.location.origin + '/test/list',
			success: function(response){
				jsonOutput = JSON.parse(response)
				console.log(jsonOutput);
				$('#contesttestlist').empty()
				if (jsonOutput.length == 0){
					$('#contesttestlist').append('<span class="titlebox">No Testcases are defined.</span>');
					$('.titlebox').css( "color", "red");
					return
				}
				$('#contesttestlist').append('<span class="titlebox">Please select testcases:</span>');
				$.each(jsonOutput, function(index, data){
					var label = $("<label></label>").addClass("container");
					label.text(data["Name"]);
					label.append(`<input type="checkbox" value="${data['Path']}">`);
					label.append('<span class="checkmark"></span>');
					$('#contesttestlist').append(label);
				});
			}
		});
	});
	$("#btnruntests").click(function(){
		console.log(mylocalStorage)
		var selectedtests = []
		var jsondata = {}
		$('#testsem100console').contents().find("head").remove();
		$('#testsem100console').contents().find("body").remove();
		$('#testsem100console').removeAttr("src");
		$('#contesttestlist input:checkbox:checked').each(function(){
			selectedtests.push($(this).val())
		});
		if ( selectedtests.length < 1 ) {
			$('.titlebox').effect( "highlight", {color:"red"}, 3000 );
			return
		}
		jsondata['testlist'] = JSON.stringify(selectedtests)
		console.log(selectedtests)
		$.ajax({
			type: "POST",
			url: window.location.origin + '/test/start/' + localStorage.getItem('username'),
			data: jsondata,
			dataType: 'html',
			success: function(response){
				console.log(response)
                                console.log("Test console started");
                                $('#testsem100console').attr("src", window.location+"console");
                                $('#dltestreport').prop("disabled", false);
			}
		});
	});
	$("#dltestreport").click(function(){
                $("#modalDownloadBody").html("Downloading your test logs ...");
                $('#p-downloadtests').css("display", "");
                $('#progress-downloadtests').css("width","0%");
                $.ajax({
                        type: "POST",
                        url: window.location.origin + '/test/logs/' + localStorage.getItem('username'),
                        contentType: 'application/octet-stream',
                        xhrFields:{
                           responseType: 'blob',
                           onprogress: function(progress){
                                   var percentage = Math.floor((progress.loaded / progress.total) * 100);
                                   $('#progress-downloadtests').css("width",percentage+"%");
                                   console.log(percentage)
                           }
                        },
                        success: function(response){
                                $("#modalDownload").modal("hide");
                                $("#modalDownloadBody").html("Download completed");
                                var link=document.createElement('a');
                                var url = window.URL || window.webkitURL;
                                link.href=url.createObjectURL(response);
                                link.download="contest_log.zip";
                                link.click();
                        }
                });
        });

});
