<template>
    <header>
        <nav>
            <div class="left"><img src="logo.svg"></div>
            <div class="right" id="app-login">
                <button @click="login" class="tertiary">Log in</button>
                <button @click="signup" class="tertiary">Sign up</button>
            </div>
        </nav>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 30" preserveAspectRatio="none">
            <polygon class="svg--sm" fill="#343434" points="0,0 10,30 40,0 50,30 70,10 90,0 100,30 100,100 0,100"/>
            <polygon class="svg--lg" fill="#343434"
                     points="0,0 10,30 20,10 40,0 50,30 70,10 80,20 90,0  100,30 100,100 0,100"/>
        </svg>
    </header>
</template>

<script lang="ts">
import { Vue, Component, Prop } from "vue-property-decorator";

@Component
export default class HelloDecorator extends Vue {
    @Prop() name!: string;
    @Prop() initialEnthusiasm!: number;

    enthusiasm = this.initialEnthusiasm;

    login() {
        this.enthusiasm++;
    }
    signup() {
        if (this.enthusiasm > 1) {
            this.enthusiasm--;
        }
    }

    get exclamationMarks(): string {
        return Array(this.enthusiasm + 1).join('!');
    }
}
</script>
