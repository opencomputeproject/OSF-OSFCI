function formSubmission(id, fn, successMsg, errorMsg) {
        // This is to detect the form submission 
        $(id).submit(function (e) {
		
                // To avoid the default submission
                e.preventDefault();

		// Looping over the form elements to build the parameter string
		var Parameters = "";
		var username = "";

		$($(id).prop('elements')).each(function(){
		    if ( this.type != "submit" )
		    {
			    if ( Parameters.length > 1 )
				    Parameters = Parameters + '&' + this.placeholder +'=' + this.value ;
			    else
				    Parameters = Parameters + this.placeholder +'=' + this.value ;
			    if ( this.placeholder == "username" )
			    {
				    mylocalStorage['username'] = this.value;
				    username = this.value;
			    }
		    }
		});
	
                var Url = '/user/'+username+'/'+fn;
                var jqxhr = $.post(Url, Parameters,
                        function postreturn(data) {
				if ( data != "" ) {
					// Did we got a JSON result ?
					// if yes we must store the value into RAM
					// We are catching up the JSON key and store the data into the global locaStorage
					try {
						var obj = JSON.parse( data );
						var myarray = Object.keys(obj);
                                	        for (let i = 0; i  < myarray.length; i++) {
                                       		         mylocalStorage[myarray[i]] = obj[myarray[i]];
                                        	}
                                        	// we are logged
                                        	logged();
					} catch(e) {
						// This is not a JSON file
						$("#formAnswer").css('color', 'red');
						$("#formAnswer").text(errorMsg);
					}
				
        			}
        			else
 			        {
			                $("#formAnswer").css('color', 'green');
			                $("#formAnswer").text(successMsg);
			                $("#btn1").hide();
        			}
			},
                        'text'
		);
		jqxhr.fail( function() {
			$("#formAnswer").css('color', 'red');
			$("#formAnswer").text("Auth Error");
		});
        });
}
