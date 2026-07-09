import * as React from 'react';
import { cn } from '@/lib/utils';

const Avatar = React.forwardRef<HTMLSpanElement, React.HTMLAttributes<HTMLSpanElement>>(
  ({ className, ...props }, ref) => (
    <span
      ref={ref}
      className={cn(
        'relative flex size-7 shrink-0 overflow-hidden rounded-full border border-border',
        className,
      )}
      {...props}
    />
  ),
);
Avatar.displayName = 'Avatar';

const AvatarImage = React.forwardRef<HTMLImageElement, React.ImgHTMLAttributes<HTMLImageElement>>(
  ({ className, alt, ...props }, ref) => (
    // eslint-disable-next-line jsx-a11y/alt-text
    <img ref={ref} alt={alt} className={cn('aspect-square size-full object-cover', className)} {...props} />
  ),
);
AvatarImage.displayName = 'AvatarImage';

export { Avatar, AvatarImage };
