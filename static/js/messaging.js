$(document).ready(function(){
    var chatbox = $("#chatbox");
    var bottom = $("#bottom");
    var cached_data_size = 0;

    function urlify(text) {
        var urlRegex = /(https?:\/\/[^\s]+)/g;
        return text.replace(urlRegex, function(url) {
            return '<a href="' + url + '"target=\"_blank\">' + url + '</a>';
        });
    }

    function validateInput(inputText) {
        return $($.parseHTML(inputText)).text();
    }

    function constructMessage(message){
        messageUser = message.Name;
        messageText = urlify(validateInput(message.Message));
        return "<div id=\"message_display\">"+messageUser+": "+messageText+"</div>";
    }

    function updateChatLog(list, cached, count){
        for(var i = cached; i < count; i++){
            chatbox.append(constructMessage(list[i]));
        }
    }
    function welcomeMessage(){
            chatbox.append("<h1>Welcome to go-webchat!</h1>");
    }

    welcomeMessage();
    setInterval(function() {
        $.get("/get_messages",function(data){
            var list = $.parseJSON(data);
            if(list.length == cached_data_size){
                return false;
            }
            updateChatLog(list, cached_data_size, list.length);
            cached_data_size = list.length;
            chatbox.scrollTop(1000);
            return false;
        });
    }, 1000);

    $("#submitmsg").click(function(){	
        var clientmsg = $("#message").val();
        if(clientmsg === "")
            return false;
        $.post("/post_message", {message: clientmsg});
        $("#message").val('');
        return false;
    });
});
