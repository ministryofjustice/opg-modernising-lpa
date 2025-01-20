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

            cy.url().then(u => {
                cy.contains('button', 'Continue').click();
                cy.visit(u.split('/').slice(3, -1).join('/') + '/task-list');
            });

            cy.contains('a', 'Confirm your identity').click();
            cy.contains('label', 'I will confirm my identity at a Post Office').click();
            cy.contains('button', 'Continue').click();

            cy.visitLpa('/progress');
            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG')
            cy.contains('You have chosen to confirm your identity at a Post Office');

            cy.contains('a', 'Go to task list').click();
            cy.contains('a', 'Confirm your identity').click();
            cy.contains('label', 'to complete my Post Office identity confirmation').click();
            cy.contains('button', 'Continue').click();
            cy.origin('http://localhost:7012', () => {
                cy.contains('button', 'Continue').click();
            });

            cy.contains('Important:').should('not.exist');
            cy.visitLpa('/progress');
        });

        it('when the lpa is submitted', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa');
            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG')
            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');
        });

        it('when more evidence is required', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&paymentTaskProgress=MoreEvidenceRequired');
            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('We need some more evidence to make a decision about your LPA fee');
            cy.contains('We contacted you on 2 April 2023 with guidance about what to do next.');
        });

        it('when the voucher has been contacted', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&idStatus=donor:insufficient-evidence');
            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('We have contacted Simone Sutherland to confirm your identity');
            cy.contains('You do not need to take any action.');
        });

        it('when the voucher has not been contacted due to outstanding payment', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&paymentTaskProgress=InProgress&idStatus=donor:insufficient-evidence');
            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('You must pay for your LPA');
            cy.contains('Return to your task list to pay for your LPA. We will then be able to contact Simone Sutherland to ask them to confirm your identity.');
        });

        it('when a vouch attempt has been unsuccessful', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&failedVouchAttempts=1&idStatus=donor:insufficient-evidence');
            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('Simone Sutherland has been unable to confirm your identity');
            cy.contains('We contacted you on 2 April 2023 with guidance about what to do next.');
        });

        it('when status is do not register', () => {
            cy.visit('/fixtures?redirect=/progress&progress=doNotRegister');
            cy.checkA11yApp();
            cy.contains('Important:');
            cy.contains('1 notification from OPG');

            cy.contains('There is a problem with your LPA');
        });

        it('when the voucher has successfully vouched - lpa not signed', () => {
            cy.visit('/fixtures?redirect=/progress&progress=confirmYourIdentity&voucher=1&idStatus=donor:vouched');
            cy.checkA11yApp();
            cy.contains('Success: 1 notification from OPG')

            cy.contains('Simone Sutherland has confirmed your identity');
            cy.contains('Return to your task list for information about what to do next.');

            cy.reload()

            cy.contains('Success: 1 notification from OPG').should('not.exist');
            cy.contains('Simone Sutherland has confirmed your identity').should('not.exist');
            cy.contains('Return to your task list for information about what to do next.').should('not.exist');
        });

        it('when the voucher has successfully vouched - lpa signed', () => {
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&idStatus=donor:vouched');
            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('Success: 1 notification from OPG')

            cy.contains('Simone Sutherland has confirmed your identity');
            cy.contains('You do not need to take any action.');

            cy.reload()

            cy.contains('Important: 1 notification from OPG');

            cy.contains('Success: 1 notification from OPG').should('not.exist');
            cy.contains('Simone Sutherland has confirmed your identity').should('not.exist');
            cy.contains('You do not need to take any action.').should('not.exist');
        });

        it("when reduced fee approved and payment task complete", () => {
            cy.visit('/fixtures?redirect=/progress&progress=payForTheLpa&feeType=NoFee');

            cy.checkA11yApp();
            cy.contains('Success: 1 notification from OPG');

            cy.contains('We have approved your LPA fee request');
            cy.contains('Your LPA is now paid.');

            cy.reload()

            cy.contains('Success: 1 notification from OPG').should('not.exist');
            cy.contains('We have approved your LPA fee request').should('not.exist');
            cy.contains('Your LPA is now paid.').should('not.exist');
        })

        it("when LPA withdrawn", () => {
            cy.visit('/fixtures?redirect=/progress&progress=withdrawn&feeType=NoFee');

            cy.checkA11yApp();

            cy.contains('Thank you for filling in your LPA').should('not.exist');

            cy.contains('Important: 1 notification from OPG');

            cy.contains('LPA revoked');
            cy.contains('We contacted you on 2 April 2023 confirming your LPA has been revoked. OPG will not register it and it cannot be used as a legal document.');

            cy.contains('Success: 1 notification from OPG').should('not.exist');
            cy.contains('We have approved your LPA fee request').should('not.exist');
            cy.contains('Your LPA is now paid.').should('not.exist');

            cy.get('#progress').should('not.exist');
            cy.contains('Thank you for filling in your LPA.').should('not.exist');
            cy.contains('Go to task list').should('not.exist');
        })
    });
});
