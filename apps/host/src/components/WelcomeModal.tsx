import { useState } from 'react';

import onboardingVideo from '~/assets/videos/video-onboarding.mp4';
import { Button } from '~/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '~/components/ui/dialog';

const WELCOME_MODAL_SEEN_KEY = 'tw_welcome_modal_seen';

export function WelcomeModal() {
  const [open, setOpen] = useState(() => {
    try {
      return window.localStorage.getItem(WELCOME_MODAL_SEEN_KEY) !== 'true';
    } catch {
      return false;
    }
  });

  const close = () => {
    setOpen(false);
    try {
      window.localStorage.setItem(WELCOME_MODAL_SEEN_KEY, 'true');
    } catch {
      return;
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        if (!open) {
          close();
        }
      }}
    >
      <DialogContent
        showCloseButton={false}
        className="tw:max-w-xs tw:gap-4 tw:rounded-3xl tw:p-6 tw:sm:max-w-md"
      >
        <video
          src={onboardingVideo}
          aria-hidden="true"
          autoPlay
          loop
          muted
          playsInline
          preload="metadata"
          className="tw:mx-auto tw:aspect-square tw:w-64 tw:rounded-2xl"
        />

        <div className="tw:flex tw:flex-col tw:gap-2">
          <DialogTitle className="tw:flex tw:items-center tw:gap-2 tw:leading-snug tw:font-semibold">
            Welcome to Teacher Workspace
            <span className="tw:rounded-full tw:bg-[#eaf3ff] tw:px-1.5 tw:py-0.5 tw:text-xs tw:font-medium tw:text-primary">
              Beta
            </span>
          </DialogTitle>

          <DialogDescription className="tw:flex tw:flex-col tw:gap-2">
            <span>
              One place for all your tools built for teachers, designed to save time and keep
              everything within reach.
            </span>

            <span>
              Early access for selected teachers. Share your thoughts via the Help icon in the
              sidebar.
            </span>
          </DialogDescription>
        </div>

        <DialogClose render={<Button className="tw:justify-self-end" />}>Get started</DialogClose>
      </DialogContent>
    </Dialog>
  );
}
