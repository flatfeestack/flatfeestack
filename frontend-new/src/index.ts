import Vue from "vue";
import VueRouter from 'vue-router';

Vue.use(VueRouter);

import {default as auth} from './auth';
import App from './components/App.vue'
import Signup from './components/Signup.vue'
import Login from './components/Login.vue'
import Dashboard from './components/Dashboard.vue'

function requireAuth (to, from, next) {
    if (!auth.loggedIn()) {
        next({
            path: '/login',
            query: { redirect: to.fullPath }
        })
    } else {
        next()
    }
}

const router = new VueRouter({
    mode: 'history',
    routes: [
        { path: '/dash', component: Dashboard, beforeEnter: requireAuth },
        { path: '/signup', component: Signup },
        { path: '/login', component: Login },
        { path: '/logout',
            beforeEnter (to, from, next) {
                auth.logout()
                next('/')
            }
        }
    ]
})

let v = new Vue({
    el: "#app",
    router,
    template: '<App/>',
    components: { App }
});
