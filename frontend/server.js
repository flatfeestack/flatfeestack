import express from 'express';

const port = process.env.PORT || 9085;
const app = express();

// serve static assets normally
app.use(express.static(process.cwd() + '/dist/client'));

// handle every other route with index.html, which will contain
// a script tag to your application's JavaScript file(s).
app.get('*', function (request, response) {
    response.sendFile(process.cwd()+ '/dist/client/index.html');
});

app.listen(port, '0.0.0.0');
console.log("server started on port " + port);