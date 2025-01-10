describe('Progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('Important:').should('not.exist');

        cy.contains('li', 'LPA paid for Not completed');
        cy.contains('li', 'Your identity confirmed Not completed');
        cy.contains('li', 'LPA signed by you Not completed');
        cy.contains('li', 'LPA certificate provided Not completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })

    it('when LPA submitted', () => {
        cy.visit('/fixtures?redirect=/progress&progress=signTheLpa');
        cy.checkA11yApp();

        cy.contains('li', 'LPA paid for Completed');
        cy.contains('li', 'Your identity confirmed Completed');
        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA certificate provided Not completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })

    it('when certificate provided', () => {
        cy.visit('/fixtures?redirect=/progress&progress=signedByCertificateProvider');
        cy.checkA11yApp();

        cy.contains('li', 'LPA paid for Completed');
        cy.contains('li', 'Your identity confirmed Completed');
        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA certificate provided Completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })

    describe('shows a notification', () => {
        it('when going to the post office', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');

            cy.contains('a', 'Confirm your identity').click();
            cy.contains('button', 'Continue').click();
            cy.go(-2);
            cy.contains('a', 'Confirm your identity').click();
            cy.contains('label', 'I will confirm my identity at a Post Office').click();
            cy.contains('button', 'Continue').click();

            cy.visitLpa('/progress');
            cy.checkA11yApp();
            cy.contains('Important:')
            cy.contains('1 notification from OPG');
            cy.contains('You have chosen to confirm your identity at a Post Office');

            cy.contains('a', 'Go to task list').click();
            cy.contains('a', 'Confirm your identity').click();
            cy.contains('label', 'to complete my Post Office identity confirmation').click();
            cy.contains('button', 'Continue').click();
            cy.contains('button', 'Continue').click();

            cy.contains('Important:').should('not.exist');
            cy.visitLpa('/progress');
        });

        it('when the lpa is submitted', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa');
            cy.checkA11yApp();
            cy.contains('Important:')
            cy.contains('1 notification from OPG');
            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');
        });

        it('when more evidence is required', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&paymentTaskProgress=MoreEvidenceRequired');
            cy.checkA11yApp();
            cy.contains('Important:')
            cy.contains('2 notifications from OPG');

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('We need some more evidence to make a decision about your LPA fee');
            cy.contains('We contacted you on 2 April 2023 at 3:04am with guidance about what to do next.');
        });

        context('when paid', () => {
            it('when the voucher has been contacted', () => {
                cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&idStatus=donor:insufficient-evidence');
                cy.checkA11yApp();
                cy.contains('Important:')
                cy.contains('2 notifications from OPG');

                cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

                cy.contains('We have contacted Simone Sutherland to confirm your identity');
                cy.contains('You do not need to take any action.');
            });
        })

        context('when not paid', () => {
            it('when the voucher has not been contacted', () => {
                cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&paymentTaskProgress=InProgress&idStatus=donor:insufficient-evidence');
                cy.checkA11yApp();
                cy.contains('Important:')
                cy.contains('2 notifications from OPG');

                cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

                cy.contains('You must pay for your LPA');
                cy.contains('Return to your task list to pay for your LPA. We will then be able to contact Simone Sutherland to ask them to confirm your identity.');
            });
        })

        it('when a vouch attempt has been unsuccessful', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&failedVouchAttempts=1&idStatus=donor:insufficient-evidence');
            cy.checkA11yApp();
            cy.contains('Important:')
            cy.contains('2 notifications from OPG');

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('Simone Sutherland has been unable to confirm your identity');
            cy.contains('We contacted you on 2 April 2023 at 3:04am with guidance about what to do next.');
        });
    })
});
