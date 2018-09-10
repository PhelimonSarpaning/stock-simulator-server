$( document ).ready(function() {


	console.log("login js");


	/*  WEBSOCKETS */

	let externalServer = "159.89.154.221";
    let localServer = window.location.host;
    let wsUri = "wss://"+ externalServer + "/ws";
    var output;
    var webSocket;
    var auth_token = {
        "uid": 0,
        "pw": "",
    };

    function init()
    {
        refreshSocket();
    }

    function refreshSocket()
    {
        webSocket = new WebSocket(wsUri);
        webSocket.onopen = function(evt) { onOpen(evt) };
        webSocket.onclose = function(evt) { onClose(evt) };
        webSocket.onmessage = function(evt) { onMessage(evt) };
        webSocket.onerror = function(evt) { onError(evt) };
    }

    function onOpen(evt)
    {
        onEvent("Connected");
        //doSend('{"container_type": "register", "register_action": "register", "device_type": "test", "device_name":"' + window.prompt("device_name","test")  + '"}');
        //doSend('{"action": "login", "msg": {"username": "Will", "password":"pass"}}');

        //doSend('{"op":"subscribe","type":"alert", "system":"irRemote"}');
        //doSend('{"action": "new_account", "msg": {"username": "Brennan", "password":"pass", "display_name":"Brennan"}}');
    }

    function onClose(evt)
    {
        onEvent("Disconnected");
    }

    function onEvent(message){
        // writeToScreen('<span style="color: darkorange;">'+ message+'</span>')
        console.log(message);
    }

    var routeLogin = function(msg) {
        console.log("login recieved");
    };


    function onMessage(evt)
    {
        
        console.log(evt.data);
        var str = JSON.stringify(JSON.parse(evt.data), undefined, 4);
        
    }

    

    function onMessage(evt)
    {
        var msg = JSON.parse(evt.data);

        var router = {
            'login': routeLogin,
            'object': routeObject,
            'update': routeUpdate,
            'alert': alertUpdate,
        }
        router[msg.action](msg);
    };

    var routeLogin = function(msg) {
        console.log("login recieved");

        // if success if true -> set cookie and forward to dashboard
        console.log(msg.msg.success);

        if(msg.msg.success) {
            // Save data to sessionStorage
            sessionStorage.setItem("authenticated", msg.msg.uuid);
            sessionStorage.setItem("auth_obj", JSON.stringify(auth_token));
            window.location.href = "/dashboard.html";

        } else {
            let err_msg = msg.msg.err;
            $('.login-err').text("Username or password is incorrect");
            //$('.login-err').text(err_msg);
        }   
        
        console.log(msg.msg.uuid);
    };

    var routeObject = function(msg) {
        switch (msg.msg.type) {
            case 'portfolio':
                
                break;

            case 'stock':
                
                break;

            case 'ledger':
                
                break;
        }
    };
    
    var routeUpdate = function(msg) {
        
    };

    var alertUpdate = function(msg) {
        
    }

    


    /////


    function doSend(message)
    {
        
        webSocket.send(message);
    }

    window.addEventListener("load", init, false);

    function attemptLogin() {
    	
    	let input_uid = $('#login-uid').val();
		let input_pw = $('#login-pw').val();
		let auth_msg = {};

    	if(input_uid != '' && input_pw != '') {
            input_uid_trimmed = input_uid.trim();
            input_pw_trimmed = input_pw.trim();
    		auth_msg = {
    					action: "login",
				        msg: {
				        	"username": input_uid_trimmed, 
				        	"password": input_pw_trimmed
				        }
				    };
            auth_token["uid"] = input_uid_trimmed;
            auth_token["pw"] = input_pw_trimmed;
    	}

    	console.log(auth_msg);

    	try {
    		
	  		doSend(JSON.stringify(auth_msg));

            //doSend('{"action": "login", "msg": {"username": "Will", "password":"pass"}}');
		}

		catch(error) {
	  		console.error(error);
		  	$('.login-err').text("Username or password is incorrect");
		}
		
    }

    $('#input-login-submit').click(function() {
    
    	attemptLogin();
        console.log("login clicked");
        
    });

    $(document).keypress(function(e) {
        
            if(e.which == 13) {
                attemptLogin();

            }
        
        });

	window.addEventListener("load", init, false);

    // setInterval(function() {

    //     doSend('{"action": "chat", "msg": {"message_body": "FUCK YEAH"}}');
        

    // }, 249);

});