<template>
    <div>
    <section class="parallax" style="text-align: center;padding-bottom: 2em">
        <div class="row" style="margin-bottom: 3em;">
            <div class=" col-sm-offset-1 col-sm-10">
                <h1>Flat-fee Open Source Software Donations</h1>
                <h2>Show your <span class="icon-heart"></span> for open source software</h2>
            </div>
        </div>

        <div class="row">
            <div class="col-sm-12 col-md-4 col-lg-4" style="border-right: solid 2px #fff">
                <h2>Become a backer</h2>
                <h3>Support your favorite stack and projects</h3>
                <p style="padding: 3vw">For a flat fee of $5, you can support as many open source project as you like. Your
                    developer badge shows which projects you are <b>passionate about</b>.</p>
            </div>

            <div class="col-sm-12 col-md-4 col-lg-4" style="border-right: solid 2px #fff">
                <h2>Monetize your project</h2>
                <h3>Put a value on your open source contribution</h3>
                <p style="padding: 3vw"><b>Earn money</b> if your open source project is backed. Your developer badge shows
                    a multiplicator,
                    earned money divided by $5. Compare your score with other developers.</p>

            </div>


            <div class="col-sm-12 col-md-4 col-lg-4">
                <h2>Become a sponsor</h2>
                <h3>Showcase your corporate technical culture</h3>
                <p style="padding: 3vw">As a company you can pay $5 for your developers, which can back their projects
                    and stacks. Your company badge shows your technical culture, what your employees are <b>passionate
                        about</b>.</p>
            </div>
        </div>

        <div class="row">
            <div class="col-sm-12 col-md-4 col-lg-4">
                <button class="primary">Become a backer</button>
            </div>


            <div class="col-sm-12  col-md-4 col-lg-4">
                <button class="primary">Earn money</button>
            </div>


            <div class="col-sm-12  col-md-4 col-lg-4">
                <button class="primary">Become a sponser</button>
            </div>

        </div>
    </section>

    <div class="row" style="background-color: darkgray;padding-bottom: 2em">
        <div class="col-sm-12 col-md-6 col-lg-6"
             style="margin-top: 1em;padding: 3vw; border-right: solid 2px #fff; height: 100%">
            <h2>What are our fees?</h2>
            <p>Since we have full-time jobs, this being a side-project, and our love for open source software, <b>we charge
                0%</b>.
                That means building this site, hosting fees, accounting, etc. is donated by us. We
                would like to distribute the full amount you payed to the projects, however, there are service fees ~5%
                (e.g.,
                credit card fees) that is deducted from your $5.</p>
            <p>&nbsp;</p>
            <div class="container">
                <div class="row">
                    <div class="col-sm-8">Your awesome projects get
                        <mark class="tertiary">95%</mark>
                    </div>
                    <div class="col-sm-4" style="text-align: right;">We keep
                        <mark class="grey">0%</mark>
                    </div>
                </div>
                <div class="row">
                    <div class="col-sm-12">
                        <progress value="95" max="100" class="tertiary" style="margin:0px;"></progress>
                    </div>
                    <div class="col-sm-12" style="text-align: right;">
                        Transaction fees average 5%<br>
                        <small>(e.g., 3rd party credit card serices)</small>
                    </div>
                </div>
            </div>
        </div>
        <div class="col-sm-12 col-md-6 col-lg-6" style="margin-top: 1em;padding: 3vw; height: 100%">
            <h2>What about Open Collective or Patreon?</h2>
            <p>Donation services support single project or person based contributions:
                e.g., support project X with $50 per year, or support person Y with $50 per year. <b>The problem is </b>
                that the more projects or people you support, the more you pay.</p>

            <p>From our personal experience, when we come across an interesting project and, we <b>often end up not
                supporting</b>, as we would have to support at least 50 other interesting projects. Paying for every
                single project or person does not work for us and we would rather have a single flat fee to support
                those 50 projects.</p>

            <p><b>The solution</b> is <b>flat fee contribution</b>. Each project receives your donation <b>equally
                divided</b>.
                Simply donate and don't worry about how many projects you can support, its always $5.
            </p>
        </div>
    </div>

    <div class="container" style="padding-top:3em;padding-bottom:3em;background-color: #fff">
        <div class="row" style="background-color: white;text-align: center;">
            <div class="col-sm-12 col-md-12 col-lg-12">
                <h2 style="margin-left: auto;margin-right: auto;">How much can you earn for a <u>single</u> project?</h2>
                <p>It is assumed that a backer backs 10 projects in average</p>
            </div>
            <div class="col-sm-12 col-md-4 col-lg-4">
                <h3>Supporters: {{useroutput}}</h3>
                <input type="range" min="0" max="1000000" value="100000" step="10000" class="slider" id="users" v-model="users">
            </div>


            <div class="col-sm-12  col-md-4 col-lg-4">
                <h3>Conversion rate: {{rateoutput}}</h3>
                <input type="range" min="0" max="120" value="50" step="10" class="slider" id="rate" v-model="rate">
            </div>


            <div class="col-sm-12  col-md-4 col-lg-4">
                <h3>Your contribution: {{contriboutput}}</h3>
                <input type="range" min="0" max="100" value="80" step="10" class="slider" id="contrib" v-model="contrib">
            </div>
        </div>
        <div class="row" style="background-color: white;text-align: center;">
            <div class="col-sm-12 col-md-6 col-lg-6">
                <h2>$<span id="earn">{{earn}}</span>/month*</h2>
                <small>*Estimate includes 5% average transaction fee.</small>
            </div>
            <div class="col-sm-12 col-md-6 col-lg-6" style="margin-top: 2em;">
                <button class="primary">Monetize your project</button>
            </div>

        </div>
    </div>



        <div style="background-color: darkgray;text-align: center;padding-bottom: 2em;;padding-top: 2em;">
            <div class="col-sm-12 col-md-12 col-lg-12">
                <h2 style="text-align: center;">Your example project badge</h2>
                <p>Make sure your visitors have the chance to see that they can support you. Show this badge on your project site. Tell me how to <a href="create">create a project badge</a>.</p>

                <!-- badge-inline.svg-->

                <h2 style="text-align: center;">Show me the other badges!</h2>
                <p>As a backer, earner, or sponser you also get a badge that you can embed into your site, personal, or company profile.</p>
                <div class="row">
                    <div class="col-sm-12 col-md-4 col-lg-4" style="border-right: solid 2px #fff">
                        <h3>Your badge as a backer</h3>

                        <!-- backer-inline.svg-->
                    </div>


                    <div class="col-sm-12  col-md-4 col-lg-4" style="border-right: solid 2px #fff">
                        <h3>Your badge as an earner</h3>

                        <!-- earner-inline.svg-->
                    </div>


                    <div class="col-sm-12  col-md-4 col-lg-4">
                        <h3>Your badge as a sponsor</h3>
                        <!-- <div ref="criteria" v-html="require('!html-loader!../../sponsor-inline2.svgi')"></div> -->

                        <!-- <object id="test" data="./sponsor-inline2.svgi" ></object> -->
                        <!-- sponsor-inline.svgi-->
                    </div>

                </div>


            </div>
        </div>




    <div class="row" style="background-color: white;">
        <div class="col-sm-12 col-md-12 col-lg-12">
            <h2 style="text-align: center;">The flatfeestack Process in 4 Steps</h2>
        </div>
    </div>
    <div class="row" style="background-color: white;">
        <div class="col-sm-12 col-md-3 col-lg-3">
            <div>
                <h2 style="text-align: center;">1. Register Project</h2>
                <p>Any backer, earner or sponsor can register a project with a public git repository. At the moment, only git is supported.</p>
            </div>
        </div>

        <div class="col-sm-12 col-md-3 col-lg-3">
            <div>
                <h2 style="text-align: center;">2. Support Project</h2>
                <p>Any backer or earner can support a project that was registered in step 1. Simply click an a badge or search the project in our database.</p>
            </div>
        </div>

        <div class="col-sm-12 col-md-3 col-lg-3">
            <div>
                <h2 style="text-align: center;">3. We Monitor Contributions</h2>
                <p>Flatfeestack monitors the public repository and stores developer contributions. A link between flatfeestack users and the
                    git user/email has to be done.</p>
            </div>
        </div>

        <div class="col-sm-12 col-md-3 col-lg-3">
            <div>
                <h2 style="text-align: center;">4. Payout</h2>
                <p>Payout is done in a 3 month intervals. Within these 3 months the payout is proportional to your contribution.</p>
            </div>
        </div>
    </div>

    <div class="row" style="background-color: darkgray;">
        <div class="col-sm-12 col-md-12 col-lg-12">
            <h2 style="text-align: center;">Who are we?</h2>

            <div class="grid-img">
                <div class="img-round">
                    <p><img src="tom.jpg" alt="Thomas Bocek" title="Thomas Bocek"/></p>
                </div>
                <div>
                    <h3>Thomas Bocek</h3>
                    <p>Distributed systems researcher, Professor at HSR Rapperswil, and passionate open source user and
                        developer, <a href="mailto:tom(at)flatfeestack.com">
                            <img src="mail.svg#text" alt="Email" title="Email">tom(at)flatfeestack.com</a>
                    </p>
                </div>
            </div>

            <div class="grid-img">
                <div>
                    <h3>Guilherme Machado Sperb</h3>
                    <p>Computer Scientist (Ph.D.), Senior Software Engineer, and passionate open source user and
                        developer, <a href="mailto:guil(at)flatfeestack.com">
                            <img src="mail.svg#text" alt="Email" title="Email">guil(at)flatfeestack.com</a>
                    </p>
                </div>
                <div class="img-round">
                    <p><img src="guil.jpg" alt="Guilherme Machado Sperb" title="Guilherme Machado Sperb"/></p>
                </div>
            </div>

            <div class="grid-img">
                <div class="img-round">
                    <p><img src="naty.jpg" alt="Natalie Eisenring" title="Natalie Eisenring"/></p>
                </div>
                <div>
                    <h3>Natalie Eisenring</h3>
                    <p>Financial expert, CEO of Coinblesk, and passionate open source user,
                        <a href="mailto:naty(at)flatfeestack.com">
                            <img src="mail.svg#text" alt="Email" title="Email">naty(at)flatfeestack.com
                        </a>
                    </p>
                </div>

            </div>
        </div>
    </div>


    <div class="grid-1" style="margin-bottom: 3em; background-color: white;">
        <div>
            <h1 style="text-align: center;">FAQ</h1>
        </div>
        <div style="margin-top: 1em;padding: 1em; border-right: solid 2px #fff; height: 100%">
            <h2>What happens if a developer does not claim his earnings?</h2>
            <p>If for example
                a backer supports 16 projects and in 1 project no develper claims his earnings, the backer will give the money to the other 15 projects. If
                within a project that gets for example 300 USD with 4 developers, and one deve1oper does not claims his earnings, the others 3 get each 100 USD.</p>

            <h2>When does the payout to a developer happen?</h2>
            <p>We have 3 month periods and if after 3 months a developer does not setup his account, the developer does not get his share.</p>

            <h2>Why is the payout in Cryptocurrency?</h2>
            <p>To keep the fees as low as possible and due to much easier integration we use either Bitcoin or Ethereum. In future we plan to support other
                crypto currencies.</p>

            <h2>How can a developer claim that he owns a username?</h2>
            <p>To claim payment, in the root repo, there needs to be a directory .flatfeestack/. Each developer commits/pushes a property file with the name of his/her
                flatfeestack UUID. For example tbocek-366bc810.prop. For GitHub we also provide a service where any GitHub user can create a pull
                request that gets accepted automatically: https://github.com/flatfeestack/users.

                The property file has the following content:
                uuid = 66cbc110-9db2-4ca3-9082-49e2a0813af7 /* mandotory */
            </p>

            <h2>How do we measure contributions?</h2>
            <p>We go a git log and add the number of lines changes: git log --after="2019-01-01" --before="2019-04-01" --pretty=%aE --numstat</p>
        </div>
    </div>



    <!--

    https://css-tricks.com/creating-non-rectangular-headers/

    #ff9d2f, #ff6126

    https://stripe.com/atlas/guides/landing-page-copy
    https://www.landingfolio.com/
    https://unbounce.com/landing-page-templates/
    https://startbootstrap.com/template-categories/landing-pages/
    https://themeforest.net/category/marketing/landing-pages
    https://colorlib.com/wp/free-bootstrap-landing-pages-templates/
    https://elements.envato.com/web-templates/landing-page-templates
    https://www.wix.com/website/templates/html/landing-pages

    -->
    </div>
</template>

<script lang="ts">
    import { Vue, Component, Watch} from "vue-property-decorator";

    //import Sponsor from './../sponsor-inline.svgi';
    @Component
    export default class Calc extends Vue {
        users = 100000;
        rate = 50;
        contrib = 80;
        get earn () : string {
            const x = (5 / 10) * this.users * (this.rate / (40 * 100)) * (this.contrib / 100);
            if(x > 1000) {
                return (x / 1000) + 'k';
            } else {
                return ((x / 10) * 10) + '';
            }
        }
        get rateoutput() : string {
            return (this.rate / 40).toFixed(2) + "%"
        }
        get useroutput() : string {
            return this.users.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        }
        get contriboutput() : string {
            return this.contrib.toString() + "%";
        }

        mounted () {

            var svg = document.getElementById('test') as HTMLObjectElement;
            (svg.contentDocument as any).init(this.$refs.criteria);

           /* this.nodeScriptReplace(this.$refs.criteria);
            console.log(this.$refs.criteria)

            let chartContainer = this.$refs.criteria as HTMLDivElement

            let str = chartContainer.innerHTML.toString();

            //this.$refs.criteria = new DOMParser().parseFromString(str, "text/xml");
            let dom = new DOMParser().parseFromString(str, "text/xml");
            this.nodeScriptReplace(dom);
            console.log(dom);*/

            //chartContainer.appendChild(dom)
        }

        nodeScriptReplace(node) {
            if ( this.nodeScriptIs(node) === true ) {
                node.parentNode.replaceChild( this.nodeScriptClone(node) , node );
                //console.log("replace!!")
            }
            else {
                var i        = 0;
                var children = node.childNodes;
                while ( i < children.length ) {
                    this.nodeScriptReplace( children[i++] );
                }
            }
            console.log("ret")
            return node;
        }
        nodeScriptIs(node) {
            console.log("tga:"+node.tagName);
            return node.tagName === 'script';
        }
        nodeScriptClone(node){
            var script  = document.createElement("script");
            script.text = node.innerHTML;
            //eval(node.innerHTML);
            /*for( var i = node.attributes.length-1; i >= 0; i-- ) {
                script.setAttribute( node.attributes[i].name, node.attributes[i].value );
                console.log("replace!!");
            }*/
            return script;
        }
    }
</script>


