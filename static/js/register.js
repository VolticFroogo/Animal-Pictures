var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    $.ajax({
        url: "/register",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            Email: $("#email").val(),
            Username: $("#username").val(),
            Password: $("#password").val(),
            CaptchaV2: grecaptcha.getResponse()
        }),
        dataType: "json",
        statusCode: {
            200: function() { // OK (successful registration).
                window.location.replace(window.location.origin + "/login/?verified=false");
            },
            400: function() { // Bad request (failed recaptcha).
                toastr["error"]("You have failed the reCAPTCHA, please try again.", "Registration Failed");
            },
            406: function() { // Not acceptable (email is invalid).
                toastr["error"]("Email is invalid.", "Registration Failed");
            },
            409: function() { // Conflict (email already in use).
                toastr["error"]("There is already an account using that email.", "Registration Failed");
            },
            500: function() { // Internal server error.
                toastr["error"]("Internal server error.", "Registration Failed");
            }
        }
    });

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

$(document).ready(function(){
    toastr.options.progressBar = true;

    $("#register-button").click(function(){
        if ($("#email").val() === "" || $("#username").val() === "" || $("#password").val() === "") {
            // Email is invalid.
            toastr["error"]("At least one field has no value.");
            return;
        }

        var regex = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
        if (!regex.test(String($("#email").val()).toLowerCase())) {
            // Email is invalid.
            toastr["error"]("Email is invalid.");
            return;
        }

        if ($("#password").val() !== $("#password-confirm").val()) {
            // Passwords don't match.
            toastr["error"]("Passwords don't match.");
            return;
        }

        toastr["info"]("Registering account.");

        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "register"}).then(function(token) {
            $.ajax({
                url: "/register",
                type: "POST",
                contentType: "application/json; charset=utf-8",
                data: JSON.stringify({
                    Email: $("#email").val(),
                    Username: $("#username").val(),
                    Password: $("#password").val(),
                    Captcha: token
                }),
                dataType: "json",
                statusCode: {
                    200: function() { // OK (successful registration).
                        window.location.replace(window.location.origin + "/login/?code=0");
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    406: function() { // Not acceptable (email is invalid).
                        toastr["error"]("Email is invalid.", "Registration Failed");
                    },
                    409: function() { // Conflict (email already in use).
                        toastr["error"]("There is already an account using that email.", "Registration Failed");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Registration Failed");
                    }
                }
            });
        });
    });
});
