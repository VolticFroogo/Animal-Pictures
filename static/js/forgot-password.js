var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    $.ajax({
        url: "/forgot-password",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            Email: $("#email").val(),
            CaptchaV2: grecaptcha.getResponse()
        }),
        dataType: "json",
        statusCode: {
            200: function() { // OK (successful registration).
                toastr["success"]("If an account is registered at that email and we haven't sent you a recovery email in the last 24 hours, we have sent an email to it.");
            },
            400: function() { // Bad request (failed recaptcha).
                toastr["error"]("You have failed the reCAPTCHA, please try again.", "Email Send Failed");
            },
            500: function() { // Internal server error.
                toastr["error"]("Internal server error.", "Email Send Failed");
            }
        }
    });

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

$(document).ready(function(){
    toastr.options.progressBar = true;

    $("#button").click(function(){
        toastr["info"]("Sending forgot password email.");

        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "forgot_password"}).then(function(token) {
            $.ajax({
                url: "/forgot-password",
                type: "POST",
                contentType: "application/json; charset=utf-8",
                data: JSON.stringify({
                    Email: $("#email").val(),
                    Captcha: token
                }),
                dataType: "json",
                statusCode: {
                    200: function() { // OK (successfully sent email; if it exists).
                        toastr["success"]("If an account is registered at that email and we haven't sent you a recovery email in the last 24 hours, we have sent an email to it.");
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Email Send Failed");
                    }
                }
            });
        });
    });
});
