let express = require('express'),
    app = express(),
    port = process.env.PORT || 9086;

app.listen(port);


app.get('/', (req, res) => {
    res.send("Test")
})

app.post('/post', (req, res) => {
    res.send("Post Test")
})

app.use(express.json());

/*
app.use(function (req, res, next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
    next();
});
*/

console.log('todo list RESTful API server started on: ' + port);

/*

import { TezosToolkit } from '@taquito/taquito';

const tezos = new TezosToolkit('https://YOUR_PREFERRED_RPC_URL');
*/
