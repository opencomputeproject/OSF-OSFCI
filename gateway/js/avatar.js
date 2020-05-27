function readURL(input) {
    if (input.files && input.files[0]) {
        var reader = new FileReader();
        reader.onload = function(e) {
            $('#imagePreview').css('background-image', 'url('+e.target.result +')');
            $('#imagePreview').hide();
            $('#imagePreview').fadeIn(650);
	    Url = '/user/' + mylocalStorage['username'] + '/updateAvatar';
	    BuildSignedAuth(Url, 'PUT' , "image/jpg", function(authString) {
	    $.ajax({
	           url: window.location.origin + '/user/' + mylocalStorage['username'] + '/updateAvatar',
	           type: 'PUT',
		   headers: {
	                "Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
	                "Content-Type" : "image/jpg",
	                "myDate" : authString['formattedDate']
                   },
	           data: e.target.result,
	           contentType: 'image/jpg',
	           success: function(response) {
        	   }
        	});
             });
        }
        reader.readAsDataURL(input.files[0]);
    }
}
$("#imageUpload").change(function() {
    readURL(this);
});

// We can initialize the content
Url ='/user/' + mylocalStorage['username'] + '/getAvatar';
BuildSignedAuth(Url, 'GET' , "application/json", function(authString) {
$.ajax({
       url: window.location.origin + '/user/' + mylocalStorage['username'] + '/getAvatar',
       type: 'GET',
       headers: {
		"Authorization": "OSF " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
		"Content-Type" : "application/json",
		"myDate" : authString['formattedDate']
                },
       success: function(response) {
			jQuery("#imagePreview").css('background-image', 'url("data:image/png;base64,' + response + '")');
       }
});
});
