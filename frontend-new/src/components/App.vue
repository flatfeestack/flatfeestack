<template>
    <div>
        <headerBar></headerBar>
        <router-view></router-view>
        <landing></landing>
    </div>
</template>

<script lang="ts">
import { Vue, Component } from "vue-property-decorator";
import Header from './Header.vue'
import Landing from './Landing.vue'

@Component({components: {headerBar: Header, landing: Landing}})
export default class App extends Vue {}
</script>

<style lang="scss">
.greeting {

}
</style>