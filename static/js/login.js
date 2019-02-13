var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    $.ajax({
        url: "/login",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            Email: $("#email").val(),
            Password: $("#password").val(),
            CaptchaV2: grecaptcha.getResponse()
        }),
        dataType: "json",
        statusCode: {
            200: function() { // OK (successful login).
                var redirect = GetURLParameter("redirect");

                if (redirect != null) {
                    window.location.replace(window.location.origin + redirect);
                } else {
                    window.location.replace("/");
                }
            },
            400: function() { // Bad Request (we failed the reCAPTCHA).
                toastr["error"]("You have failed the reCAPTCHA, please try again.", "Login Failed");
            },
            401: function() { // Unauthorized (invalid login credentials).
                toastr["error"]("Invalid login credentials.", "Login Failed");
            },
            403: function() { // Forbidden (email not verified).
                toastr["error"]("You haven't verified your email yet, please check your inbox (even spam folder) to complete the registration process.", "Login Failed");
            },
            500: function() { // Internal server error.
                toastr["error"]("Internal server error.", "Login Failed");
            }
        }
    });

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

$(document).ready(function(){
    toastr.options.progressBar = true;

    var code = GetURLParameter("code");
    switch (code) {
        case "0":
            // User has just finished the registration page.
            toastr["info"]("Successfully registered account, please check your email inbox for a verification email.");
            break;
        case "1":
            // User has just verified their account via an email.
            toastr["info"]("Successfully verified email, you may now log in.");
            break;
        case "2":
            // User has clicked on the verify link but was already verified previously.
            toastr["warning"]("You were already verified before attempting to verify again.");
            break;
        case "3":
            // User has just reset their password.
            toastr["info"]("Successfully reset password, you may now log in.");
            break;
    }

    $("#login-button").click(function(){
        toastr["info"]("Logging in.");

        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "login"}).then(function(token) {
            $.ajax({
                url: "/login",
                type: "POST",
                contentType: "application/json; charset=utf-8",
                data: JSON.stringify({
                    Email: $("#email").val(),
                    Password: $("#password").val(),
                    Captcha: token
                }),
                dataType: "json",
                statusCode: {
                    200: function() { // OK (successful login).
                        var redirect = GetURLParameter("redirect");

                        if (redirect != null) {
                            window.location.replace(window.location.origin + redirect);
                        } else {
                            window.location.replace("/");
                        }
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    401: function() { // Unauthorized (invalid login credentials).
                        toastr["error"]("Invalid login credentials.", "Login Failed");
                    },
                    403: function() { // Forbidden (email not verified).
                        toastr["error"]("You haven't verified your email yet, please check your inbox (even spam folder) to complete the registration process.", "Login Failed");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Login Failed");
                    }
                }
            });
        });
    });
});
