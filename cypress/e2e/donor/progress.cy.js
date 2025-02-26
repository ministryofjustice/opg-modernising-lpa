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
            cy.contains('We contacted you on 2 January 2023 with guidance about what to do next.');
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
            cy.visit('/fixtures?redirect=/progress&progress=signTheLpa&voucher=1&vouchAttempts=1&idStatus=donor:insufficient-evidence');
            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG')

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian');

            cy.contains('Simone Sutherland has been unable to confirm your identity');
            cy.contains('We contacted you on 2 January 2023 with guidance about what to do next.');
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
            cy.contains('We contacted you on 2 January 2023 confirming your LPA has been revoked. OPG will not register it and it cannot be used as a legal document.');

            cy.contains('Success: 1 notification from OPG').should('not.exist');
            cy.contains('We have approved your LPA fee request').should('not.exist');
            cy.contains('Your LPA is now paid.').should('not.exist');

            cy.get('#progress').should('not.exist');
            cy.contains('Thank you for filling in your LPA.').should('not.exist');
            cy.contains('Go to task list').should('not.exist');
        })

        it("when six months after signing and identification not confirmed", () => {
            cy.visit(`/fixtures?redirect=/progress&progress=signTheLpa&idStatus=donor:insufficient-evidence&donor=signature-expired`);

            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG');

            cy.contains('You’ve submitted your LPA to the Office of the Public Guardian (OPG)');

            cy.contains('Your LPA cannot be registered by the Office of the Public Guardian (OPG)');
            cy.contains('You did not confirm your identity within 6 months of signing your LPA, so OPG cannot register it.');
        })

        it("when not signed and identity confirmation expired", () => {
            cy.visit(`/fixtures?redirect=/progress&progress=confirmYourIdentity&idStatus=donor:expired`);

            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG');

            cy.contains('You must confirm your identity again');
            cy.contains('You did not sign your LPA within 6 months of confirming your identity, so your identity check has expired.');
        })

        it('when statutory waiting period', () => {
            cy.visit(`/fixtures?redirect=/progress&progress=statutoryWaitingPeriod`);

            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG');

            cy.contains('Your LPA is awaiting registration');
            cy.contains('at the end of our statutory 4-week waiting period');
        });

        it('when continue with identity mismatch', () => {
            cy.visit(`/fixtures?redirect=/task-list&progress=payForTheLpa`);

            cy.contains('a', 'Confirm your identity').click();
            cy.contains('button', 'Continue').click();
            cy.origin('http://localhost:7012', () => {
                cy.contains('label', 'Charlie Cooper').click();
                cy.contains('button', 'Continue').click();
            });

            cy.contains('label', 'No').click();
            cy.contains('button', 'Continue').click();

            cy.visitLpa('/progress');

            cy.checkA11yApp();
            cy.contains('Important: 1 notification from OPG');

            cy.contains('Confirmation of identity pending');
            cy.contains('You do not need to take any action');
        });

        it('when identity mismatch resolved as immaterial', () => {
            cy.visit(`/fixtures?redirect=/task-list&progress=signTheLpa&idStatus=donor:mismatch`);
            cy.visitLpa('/progress');

            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG');

            cy.contains('Confirmation of identity pending');
            cy.contains('You do not need to take any action');

            cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
                .invoke('text')
                .then((uid) => {
                    cy.request({
                        method: 'POST',
                        url: 'http://localhost:9001/emit/opg.poas.sirius/immaterial-change-confirmed',
                        body: {
                            uid: uid.trim(),
                            actorType: 'donor',
                            actorUid: 'abc',
                        },
                    }).then((response) => {
                        expect(response.status).to.eq(200);

                        cy.visitLpa('/progress')
                        cy.waitForTextByReloading('main', 'Success: 1 notification from OPG')

                        cy.checkA11yApp();

                        cy.contains('Your identity has been confirmed');
                        cy.contains('You do not need to take any action.');

                        cy.reload()

                        cy.contains('Success: 1 notification from OPG').should('not.exist');
                    });
                });
        });

        it('when identity mismatch resolved as material', () => {
            cy.visit(`/fixtures?redirect=/task-list&progress=signTheLpa&idStatus=donor:mismatch`);
            cy.visitLpa('/progress');

            cy.checkA11yApp();
            cy.contains('Important: 2 notifications from OPG');

            cy.contains('Confirmation of identity pending');
            cy.contains('You do not need to take any action');

            cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
                .invoke('text')
                .then((uid) => {
                    cy.request({
                        method: 'POST',
                        url: 'http://localhost:9001/emit/opg.poas.sirius/material-change-confirmed',
                        body: {
                            uid: uid.trim(),
                            actorType: 'donor',
                            actorUid: 'abc',
                        },
                    }).then((response) => {
                        expect(response.status).to.eq(200);

                        cy.visitLpa('/progress')
                        cy.waitForTextByReloading('main', 'Your LPA cannot be registered by the Office of the Public Guardian (OPG)')

                        cy.checkA11yApp();

                        cy.reload()

                        const date = new Date().toLocaleDateString('en-GB', {
                            day: 'numeric',
                            month: 'long',
                            year: 'numeric'
                        });

                        cy.contains(`We contacted you on ${date} with guidance about what to do next.`);
                    });
                });
        });
    });
});
