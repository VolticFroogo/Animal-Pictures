var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    $.ajax({
        url: "/password-recovery",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            Code: GetURLParameter("code"),
            Password: $("#password").val(),
            CaptchaV2: grecaptcha.getResponse()
        }),
        dataType: "json",
        statusCode: {
            200: function() { // OK (successfully reset password).
                window.location.replace(window.location.origin + "/login/?code=3");
            },
            400: function() { // Bad request (failed recaptcha).
                toastr["error"]("You have failed the reCAPTCHA, please try again.", "Password Recovery Failed");
            },
            500: function() { // Internal server error.
                toastr["error"]("Internal server error.", "Password Recovery Failed");
            }
        }
    });

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

$(document).ready(function(){
    toastr.options.progressBar = true;

    $("#button").click(function(){
        if ($("#password").val() !== $("#confirm-password").val()) {
            toastr["error"]("Passwords are different.");
            return;
        }

        toastr["info"]("Resetting password.");

        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "reset_password"}).then(function(token) {
            $.ajax({
                url: "/password-recovery",
                type: "POST",
                contentType: "application/json; charset=utf-8",
                data: JSON.stringify({
                    Code: GetURLParameter("code"),
                    Password: $("#password").val(),
                    Captcha: token
                }),
                dataType: "json",
                statusCode: {
                    200: function() { // OK (successfully reset password).
                        window.location.replace(window.location.origin + "/login/?code=3");
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Password Recovery Failed");
                    }
                }
            });
        });

        grecaptcha.reset(); // Reset the recaptcha
    });
});
