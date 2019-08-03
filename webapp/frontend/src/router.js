import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'
import Reservation from './views/Reservation.vue'
import Trains from './views/Trains.vue'

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
      path: '/reservation',
      name: 'reservation',
      // route level code-splitting
      // this generates a separate chunk (about.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: Reservation
    },
    {
      path: '/reservation/trains',
      name: 'trains',
      // route level code-splitting
      // this generates a separate chunk (about.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: Trains
    }
  ]
})
