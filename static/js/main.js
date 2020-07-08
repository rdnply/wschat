var globalVarForWS = null;
var userLogin = null;

function getLoginFromParams() {
    var url = new URL(location.href);
    var login = url.searchParams.get("login");

    return login;
}

function getDestination() {
    var login = document.querySelector('.chatName');
    login = login.innerText;
    if (login === "General chat") {
        return "";
    }

    return login;
}

function createChatInfo(user) {
    var divChatInfo = document.createElement('div');
    divChatInfo.classList.add('chatInfo');

    var divImage = document.createElement('div');
    divImage.classList.add('image');
    divChatInfo.append(divImage);

    var pName = document.createElement('p');
    pName.classList.add('name');
    pName.innerText = user.login;
    divChatInfo.append(pName);

    var pMessage = document.createElement('p');
    pMessage.classList.add('message');
    divChatInfo.append(pMessage);


    return divChatInfo;
}

function createChatButton(user) {
    var divChatButton = document.createElement('div');
    divChatButton.classList.add('chatButton');
    divChatButton.onclick = function () {
        startChat(this);
    };

    divChatButton.append(createChatInfo(user));

    return divChatButton;
}

function addChatButton(user) {
    var chats = document.querySelector('.chats');
    var chatButton = createChatButton(user);
    chats.append(chatButton);
}

function createSentMessage(msg) {
    var div = document.createElement('div');
    div.classList.add('msg');
    div.classList.add('messageSent');
    div.innerHTML = msg;

    var i = document.createElement('i');
    div.append(i);

    var span = document.createElement('span');
    div.append(span);

    return div;
}

function addSentMessage(msgText) {
    var sentMsg = createSentMessage(msgText);

    var convHistory = document.querySelector('.convHistory');
    convHistory.append(sentMsg);
}

function sendMessage() {
    event.preventDefault();
    var input = document.querySelector('.replyMessage');
    var msgText = input.value;
    input.value = '';
    addSentMessage(msgText);

    var destination = getDestination();
    console.log('destination= ' + destination)

    globalVarForWS.send(JSON.stringify({
        from: userLogin,
        to: destination,
        message: msgText
    }));
}

function createReceivedMessage(msg) {
    var div = document.createElement('div');
    div.classList.add('msg');
    div.classList.add('messageReceived');
    div.innerHTML = msg.message;

    var span = document.createElement('span');
    span.classList.add('userName');
    span.innerHTML = msg.from;
    div.append(span);

    return div;
}

function addReceivedMessage(msg) {
    var receivedMsg = createReceivedMessage(msg);
    var convHistory = document.querySelector('.convHistory');
    convHistory.append(receivedMsg);
}

var HttpClient = function () {
    this.get = function (aUrl, aCallback) {
        var anHttpRequest = new XMLHttpRequest();
        anHttpRequest.onreadystatechange = function () {
            if (anHttpRequest.readyState === 4 && anHttpRequest.status === 200)
                aCallback(anHttpRequest.responseText);
        }

        anHttpRequest.open("GET", aUrl, true);
        anHttpRequest.send(null);
    }
}

function startChat(el) {
    var companion = el.querySelector('.name');
    companion = companion.innerText;
    var login = getLoginFromParams();
    console.log('companion= ' + companion);
    console.log('login= ' + login);
    var chatName = document.querySelector('.chatName');
    chatName.innerText = companion;

    var client = new HttpClient();
    client.get('http://localhost:5000/chat/messages?login=' + login + '&companion=' + companion, function (response) {
        console.log('response= ' + response);
        var json = JSON.parse(response);
        console.log('json= ' + json);
        var convHistory = document.querySelector('.convHistory');
        convHistory.innerHTML = '';

        if (json !== null) {
            for (const msg of json) {
                if (msg.from === login) {
                    addSentMessage(msg.message);
                } else {
                    addReceivedMessage(msg);
                }
            }
        }
    });

}