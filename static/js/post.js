var vote = false;

var recaptchaCallback = function() {
    // User has completed v2 reCAPTCHA to prove they're not a robot.
    toastr["info"]("reCAPTCHA completed, trying again.");

    sendVote(vote, true)

    $("#recaptcha-modal").modal("hide");
    grecaptcha.reset(); // Reset the reCAPTCHA.
};

var sendVote = function(upvote, useV2) {
    if (useV2) {
        $.ajax({
            url: window.location.pathname + "/vote",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                Upvote: upvote,
                CaptchaV2: grecaptcha.getResponse()
            }),
            dataType: "json",
            statusCode: {
                200: function(r) {
                    if (VoteStatus === 0) {
                        if (upvote) {
                            $("#upvote").addClass("current-vote");
                            VoteStatus = 1;
                        } else {
                            $("#downvote").addClass("current-vote");
                            VoteStatus = 2;
                        }
                    } else if (VoteStatus === 1) {
                        if (upvote) {
                            $("#upvote").removeClass("current-vote");
                            VoteStatus = 0;
                        } else {
                            $("#downvote").addClass("current-vote");
                            $("#upvote").removeClass("current-vote");
                            VoteStatus = 2;
                        }
                    } else {
                        if (upvote) {
                            $("#upvote").addClass("current-vote");
                            $("#downvote").removeClass("current-vote");
                            VoteStatus = 1;
                        } else {
                            $("#downvote").removeClass("current-vote");
                            VoteStatus = 0;
                        }
                    }

                    $("#score").text(r.Score);
                },
                400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                    toastr["error"]("You failed the reCAPTCHA.", "Vote Failed");
                },
                410: function() { // Gone (post deleted).
                    toastr["error"]("This post has been deleted.", "Vote Failed");
                },
                500: function() { // Internal server error.
                    toastr["error"]("Internal server error.", "Vote Failed");
                }
            }
        });

        $("#recaptcha-modal").modal("hide");
        grecaptcha.reset(); // Reset the reCAPTCHA.
    } else {
        grecaptcha.execute("6Lfyi5AUAAAAAJhGIO45QyuAD7L_yqIq5s0Kc6NN", {action: "vote"}).then(function(token) {
            $.ajax({
                url: window.location.pathname + "/vote",
                type: "POST",
                contentType: "application/json; charset=utf-8",
                data: JSON.stringify({
                    Upvote: upvote,
                    Captcha: token
                }),
                dataType: "json",
                statusCode: {
                    200: function(r) {
                        if (VoteStatus === 0) {
                            if (upvote) {
                                $("#upvote").addClass("current-vote");
                                VoteStatus = 1;
                            } else {
                                $("#downvote").addClass("current-vote");
                                VoteStatus = 2;
                            }
                        } else if (VoteStatus === 1) {
                            if (upvote) {
                                $("#upvote").removeClass("current-vote");
                                VoteStatus = 0;
                            } else {
                                $("#downvote").addClass("current-vote");
                                $("#upvote").removeClass("current-vote");
                                VoteStatus = 2;
                            }
                        } else {
                            if (upvote) {
                                $("#upvote").addClass("current-vote");
                                $("#downvote").removeClass("current-vote");
                                VoteStatus = 1;
                            } else {
                                $("#downvote").removeClass("current-vote");
                                VoteStatus = 0;
                            }
                        }

                        $("#score").text(r.Score);
                    },
                    400: function() { // Bad Request (we aren't trusted; fill in reCAPTCHA v2).
                        vote = upvote;
                        toastr["warning"]("Our system suspects you of being a bot, please complete the reCAPTCHA.", "Anti-Bot Verification");
                        $("#recaptcha-modal").modal("show");
                    },
                    410: function() { // Gone (post deleted).
                        toastr["error"]("This post has been deleted.", "Vote Failed");
                    },
                    500: function() { // Internal server error.
                        toastr["error"]("Internal server error.", "Vote Failed");
                    }
                }
            });
        });
    }
}

$(document).ready(function(){
    toastr.options.progressBar = true;

    $("#upvote").click(function(){
        if (LoggedIn) {
            sendVote(true, false);
        } else {
            window.location.replace(window.location.origin + "/login/?redirect=" + window.location.pathname);
        }
    });

    $("#downvote").click(function(){
        if (LoggedIn) {
            sendVote(false, false);
        } else {
           window.location.replace(window.location.origin + "/login/?redirect=" + window.location.pathname);
        }
    });
});
