$(document).ready(function(){
    $("#login").click(function(){
        var username = $("#username").val();
        var password = $("#password").val();
        // Checking for blank fields.
        if( username ==='' || password ===''){
            alert("Please fill all fields!");
        } else {
            var postData = JSON.stringify({
                "username": username,
                "password": password
              });
            var settings = {
                "async": true,
                "crossDomain": true,
                "url": "/login",
                "method": "POST",
                "headers": {
                  "Content-Type": "application/json",
                  "Cache-Control": "no-cache"
                },
                "processData": false,
                "data": postData
              }
              
              $.ajax(settings).done(function (response) {
                  if (response.token !== '') {
                        localStorage.setItem('access_token', response.token);
                        window.location = "dashboard.html";
                  }
              });
        }
    });
});