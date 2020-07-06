var globalVarForWS = null;
var userLogin = null;

function getLoginFromParams() {
    var url = new URL(location.href);
    var login = url.searchParams.get("login");

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

function sendMessage() {
    event.preventDefault();
    var input = document.querySelector('.replyMessage');
    var msgText = input.value;
    input.value = '';
    var sentMsg = createSentMessage(msgText);


    var convHistory = document.querySelector('.convHistory');
    convHistory.append(sentMsg);

    globalVarForWS.send(JSON.stringify({
        from: userLogin,
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

function receiveMessage(msg) {
    var receivedMsg = createReceivedMessage(msg);
    var convHistory = document.querySelector('.convHistory');
    convHistory.append(receivedMsg);
}

function startChat() {
    var login = this.querySelector('.name');

}