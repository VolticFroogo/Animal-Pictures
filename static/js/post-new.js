var form;

var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    var formData = new FormData(form);

    formData.append("captchaV2", grecaptcha.getResponse());

    $.ajax({
        type: "POST",
        url: "/post/new",
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        statusCode: {
            200: function(rRaw) { // OK (successfully created post).
                var r = JSON.parse(rRaw);
                window.location.replace(window.location.origin + "/post/" + r.UUID);
            },
            400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                $("#recaptcha-modal").modal("show");
            },
            413: function() { // Request entity too large (the image we attempted to upload was rejected for being too big).
                toastr["error"]("You can not upload an image over 5MB.", "Post Creation Failed");
            },
            415: function() { // Request entity too large (the image we attempted to upload was rejected for being too big).
                toastr["error"]("The file you have selected is not an image.", "Post Creation Failed");
            },
            500: function() { // Internal server error.
                toastr["error"]("Internal server error.", "Post Creation Failed");
            }
        }
    });

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

$(document).ready(function(){
    toastr.options.progressBar = true;

    $("#image-button").click(function(event){
        event.preventDefault();
        $("#image").trigger("click");
    });

    $("#submit-button").click(function(event){
        event.preventDefault();

        if (typeof $("#image")[0].files[0] === "undefined") {
            toastr["error"]("You need to select an image...");
            return;
        }

        toastr["info"]("Creating new post.");

        form = $(this).parents("form")[0];
        var formData = new FormData(form);

        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "post_new"}).then(function(token) {
            formData.append("captcha", token);

            $.ajax({
                type: "POST",
                url: "/post/new",
                data: formData,
                cache: false,
                contentType: false,
                processData: false,
                statusCode: {
                    200: function(rRaw) { // OK (successfully created post).
                        var r = JSON.parse(rRaw);
                        window.location.replace(window.location.origin + "/post/" + r.UUID);
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    413: function() { // Request entity too large (the image we attempted to upload was rejected for being too big).
                        toastr["error"]("You can not upload an image over 5MB.", "Post Creation Failed");
                    },
                    415: function() { // Request entity too large (the image we attempted to upload was rejected for being too big).
                        toastr["error"]("The file you have selected is not an image.", "Post Creation Failed");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Post Creation Failed");
                    }
                }
            });
        });
    });
});
