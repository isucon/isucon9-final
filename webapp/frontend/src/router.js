import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'
import Register from './views/Register.vue'
import Login from './views/Login.vue'
import Search from './views/Search.vue'
import Trains from './views/Trains.vue'
import Seats from './views/Seats.vue'
import Payment from './views/Payment.vue'
import Reservations from './views/Reservations.vue'
import Test from './views/Test.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/test',
      name: 'test',
      component: Test
    },
    {
      path: '/register',
      name: 'register',
      component: Register
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    },
    {
      path: '/reservation',
      name: 'search',
      component: Search
    },
    {
      path: '/reservation/trains',
      name: 'trains',
      component: Trains
    },
    {
      path: '/reservation/seats',
      name: 'seats',
      component: Seats
    },
    {
      path: '/reservation/payment',
      name: 'payment',
      component: Payment
    },
    {
      path: '/mypage/reservations',
      name: 'reservations',
      component: Reservations
    },
  ]
})
